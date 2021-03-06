#!/bin/bash
#
# Copyright 2018-2020, CS Systemes d'Information, http://www.c-s.fr
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Script customized for {{.ProviderName}} driver

{{.Header}}

print_error() {
  read line file <<<$(caller)
  echo "An error occurred in line $line of file $file:" "{"$(sed "${line}q;d" "$file")"}" >&2
  echo -n "2,${LINUX_KIND},${FULL_VERSION_ID},$(hostname),$(date +%Y/%m/%d-%H:%M:%S)" >/opt/safescale/var/state/user_data.phase1.done
}
trap print_error ERR

fail() {
  echo "PROVISIONING_ERROR: $1"
  echo -n "$1,${LINUX_KIND},${FULL_VERSION_ID},$(hostname),$(date +%Y/%m/%d-%H:%M:%S)" >/opt/safescale/var/state/user_data.phase1.done
  set +x
  exit $1
}

mkdir -p /opt/safescale/etc /opt/safescale/bin &>/dev/null
mkdir -p /opt/safescale/var/log &>/dev/null
mkdir -p /opt/safescale/var/run /opt/safescale/var/state /opt/safescale/var/tmp &>/dev/null

exec 1<&-
exec 2<&-
exec 1<>/opt/safescale/var/log/user_data.phase1.log
exec 2>&1
set -x

LINUX_KIND=
VERSION_ID=
FULL_VERSION_ID=

sfDetectFacts() {
  [[ -f /etc/os-release ]] && {
    . /etc/os-release
    LINUX_KIND=$ID
    FULL_HOSTNAME=$VERSION_ID
  } || {
    which lsb_release &>/dev/null && {
      LINUX_KIND=$(lsb_release -is)
      LINUX_KIND=${LINUX_KIND,,}
      VERSION_ID=$(lsb_release -rs | cut -d. -f1)
      FULL_VERSION_ID=$(lsb_release -rs)
    } || {
      [[ -f /etc/redhat-release ]] && {
        LINUX_KIND=$(cat /etc/redhat-release | cut -d' ' -f1)
        LINUX_KIND=${LINUX_KIND,,}
        VERSION_ID=$(cat /etc/redhat-release | cut -d' ' -f3 | cut -d. -f1)
        FULL_VERSION_ID=$(cat /etc/redhat-release | cut -d' ' -f3)
        case $VERSION_ID in
        '' | *[!0-9]*)
          VERSION_ID=$(cat /etc/redhat-release | cut -d' ' -f4 | cut -d. -f1)
          FULL_VERSION_ID=$(cat /etc/redhat-release | cut -d' ' -f4)
          ;;
        *) ;;

        esac
      }
    }
  }
}
sfDetectFacts

create_user() {
  echo "Creating user {{.User}}..."
  useradd {{.User}} --home-dir /home/{{.User}} --shell /bin/bash --comment "" --create-home || true
  echo "{{.User}}:{{.Password}}" | chpasswd
  groupadd -r docker
  usermod -aG docker {{.User}}
  SUDOERS_FILE=/etc/sudoers.d/{{.User}}
  [ ! -d "$(dirname $SUDOERS_FILE)" ] && SUDOERS_FILE=/etc/sudoers
  cat >>$SUDOERS_FILE <<-'EOF'
Defaults:{{.User}} !requiretty
{{.User}} ALL=(ALL) NOPASSWD:ALL
EOF

  mkdir /home/{{.User}}/.ssh
  echo "{{.PublicKey}}" >>/home/{{.User}}/.ssh/authorized_keys
  echo "{{.PrivateKey}}" >/home/{{.User}}/.ssh/id_rsa
  chmod 0700 /home/{{.User}}/.ssh
  chmod -R 0600 /home/{{.User}}/.ssh/*

  cat >>/home/{{.User}}/.bashrc <<-'EOF'
pathremove() {
		local IFS=':'
		local NEWPATH
		local DIR
		local PATHVARIABLE=${2:-PATH}
		for DIR in ${!PATHVARIABLE} ; do
				if [ "$DIR" != "$1" ] ; then
					NEWPATH=${NEWPATH:+$NEWPATH:}$DIR
				fi
		done
		export $PATHVARIABLE="$NEWPATH"
}
pathprepend() {
		pathremove $1 $2
		local PATHVARIABLE=${2:-PATH}
		export $PATHVARIABLE="$1${!PATHVARIABLE:+:${!PATHVARIABLE}}"
}
pathappend() {
		pathremove $1 $2
		local PATHVARIABLE=${2:-PATH}
		export $PATHVARIABLE="${!PATHVARIABLE:+${!PATHVARIABLE}:}$1"
}
pathprepend $HOME/.local/bin
pathappend /opt/safescale/bin
EOF

  chown -R {{.User}}:{{.User}} /opt/safescale
  chmod -R 0640 /opt/safescale
  find /opt/safescale -type d -exec chmod a+rx {} \;
  chmod 1777 /opt/safescale/var/tmp

  chown -R {{.User}}:{{.User}} /home/{{.User}}

  for i in /home/{{.User}}/.hushlogin /home/{{.User}}/.cloud-warnings.skip; do
    touch $i
    chown root:{{.User}} $i
    chmod ug+r-wx,o-rwx $i
  done

  echo done
}

# Follows the CentOS rules:
# - /etc/hostname contains short hostname
put_hostname_in_hosts() {
  FULL_HOSTNAME="{{ .HostName }}"
  SHORT_HOSTNAME="${FULL_HOSTNAME%%.*}"

  echo "${SHORT_HOSTNAME}" >/etc/hostname
  hostname "${SHORT_HOSTNAME}"
}

# Disable cloud-init automatic network configuration to be sure our configuration won't be replaced
disable_cloudinit_network_autoconf() {
  fname=/etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
  mkdir -p $(dirname $fname)
  echo "network: {config: disabled}" >$fname
}

disable_services() {
  case $LINUX_KIND in
  debian | ubuntu)
    if [[ -n $(which systemctl) ]]; then
      systemctl stop apt-daily.service &>/dev/null
      systemctl kill --kill-who=all apt-daily.service &>/dev/null
    fi
    if [[ -n $(which system) ]]; then
      which system && service stop apt-daily.service &>/dev/null
    fi
    ;;
  esac
}

is_network_reachable() {
  NETROUNDS=24
  REACHED=0

  for i in $(seq ${NETROUNDS}); do
    if which curl; then
      curl -I www.google.com -m 5 | grep "200 OK" && REACHED=1 && break
    fi

    if which wget; then
      wget -T 10 -O /dev/null www.google.com &>/dev/null && REACHED=1 && break
    else
      ping -n -c1 -w10 -i5 www.google.com && REACHED=1 && break
    fi

    sleep 1
  done

  if [[ ${REACHED} -eq 0 ]]; then
    echo "Unable to reach network"
    return 1
  fi

  return 0
}

# If host isn't a gateway, we need to configure temporarily and manually gateway on private hosts to be able to update packages
ensure_network_connectivity() {
  op=-1
  is_network_reachable && op=$? || true
  if [[ ${op} -eq 0 ]]; then
    echo "ensure_network_connectivity started WITH network..."
  else
    echo "ensure_network_connectivity started WITHOUT network..."
  fi

  {{- if .AddGateway }}
  route del -net default &>/dev/null
  route add -net default gw {{ .DefaultRouteIP }}
  {{- else }}
  :
  {{- end}}

  op=-1
  is_network_reachable && op=$? || true
  if [[ ${op} -eq 0 ]]; then
    echo "ensure_network_connectivity finished WITH network..."
    return 0
  else
    echo "ensure_network_connectivity finished WITHOUT network..."
  fi

  echo "" >> /etc/resolv.conf
  echo "nameserver 1.1.1.1" >> /etc/resolv.conf

  op=-1
  is_network_reachable && op=$? || true
  if [[ ${op} -eq 0 ]]; then
    echo "ensure_network_connectivity finished WITH network AFTER putting custom DNS..."
    return 0
  else
    echo "ensure_network_connectivity finished WITHOUT network, not even custom DNS was enough..."
    return 1
  fi
}

function fail_fast_unsupported_distros() {
  case $LINUX_KIND in
  debian)
    lsb_release -rs | grep "8." && {
      echo "PROVISIONING_ERROR: Unsupported Linux distribution (docker) '$LINUX_KIND $(lsb_release -rs)'!"
      fail 199
    } || true
    ;;
  ubuntu)
    if [[ $(lsb_release -rs | cut -d. -f1) -le 17 ]]; then
      if [[ $(lsb_release -rs | cut -d. -f1) -ne 16 ]]; then
        echo "PROVISIONING_ERROR: Unsupported Linux distribution '$LINUX_KIND $(lsb_release -rs)'!"
        fail 199
      fi
    fi
    ;;
  redhat | rhel | centos)
    if [[ -n $(which lsb_release) ]]; then
      if [[ $(lsb_release -rs | cut -d. -f1) -lt 7 ]]; then
        echo "PROVISIONING_ERROR: Unsupported Linux distribution (firewalld) '$LINUX_KIND $(lsb_release -rs)'!"
        fail 199
      fi
    else
      if [[ $(echo ${VERSION_ID}) -lt 7 ]]; then
        echo "PROVISIONING_ERROR: Unsupported Linux distribution (firewalld) '$LINUX_KIND $VERSION_ID'!"
        fail 199
      fi
    fi
    ;;
  fedora)
    if [[ -n $(which lsb_release) ]]; then
      if [[ $(lsb_release -rs | cut -d. -f1) -lt 30 ]]; then
        echo "PROVISIONING_ERROR: Unsupported Linux distribution '$LINUX_KIND $(lsb_release -rs)'!"
        fail 199
      fi
    else
      if [[ $(echo ${VERSION_ID}) -lt 30 ]]; then
        echo "PROVISIONING_ERROR: Unsupported Linux distribution '$LINUX_KIND $VERSION_ID'!"
        fail 199
      fi
    fi
    ;;
  *)
    echo "PROVISIONING_ERROR: Unsupported Linux distribution '$LINUX_KIND $(lsb_release -rs)'!"
    fail 199
    ;;
  esac
}

# ---- Main

export DEBIAN_FRONTEND=noninteractive

put_hostname_in_hosts
disable_cloudinit_network_autoconf
disable_services
create_user
ensure_network_connectivity || true

touch /etc/cloud/cloud-init.disabled

fail_fast_unsupported_distros

echo -n "0,linux,${LINUX_KIND},${FULL_VERSION_ID},$(hostname),$(date +%Y/%m/%d-%H:%M:%S)" >/opt/safescale/var/state/user_data.phase1.done
set +x
exit 0
