# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr
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

#### Installs and configure common tools for any kind of nodes ####

install_common_requirements() {
    echo "Installing common requirements..."

    export LANG=C

    # Creates user cladm
    useradd -s /bin/bash -m -d /home/cladm cladm
    groupadd -r -f docker &>/dev/null
    usermod -aG docker gpac
    usermod -aG docker cladm
    echo "cladm:{{ .CladmPassword }}" | chpasswd
    mkdir -p /home/cladm/.ssh && chmod 0700 /home/cladm/.ssh
    mkdir -p /home/cladm/.local/bin && find /home/cladm/.local -exec chmod 0770 {} \;
    cat >>/home/cladm/.bashrc <<-'EOF'
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
pathappend /opt/mesosphere/bin
EOF
    chown -R cladm:cladm /home/cladm
}
export -f install_common_requirements

case $LINUX_KIND in
    centos|redhat)
        yum makecache fast
        yum install -y curl wget time jq rclone
        ;;
    debian|ubuntu)
        wait_for_apt && apt update && \
        wait_for_apt && apt install -y curl wget time jq rclone
        ;;
    *)
        echo "unmanaged Linux distribution '$LINUX_KIND'"
        exit 1
esac

/usr/bin/time -p bash -c install_common_requirements
