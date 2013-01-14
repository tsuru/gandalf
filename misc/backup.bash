#!/bin/bash -e

# Copyright 2013 tsuru authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# This script is used to backup on s3 repositories created by gandalf.

function send_to_s3 {
    echo "Sending $1 to $2 in s3 ..."
    s3cmd cp $1 $2
}

function compact {
    echo "Compacting $1 ..."
    tar zcvf $1.tar.gz $1
}

# making the backup for authorized_keys
[ -f "${HOME}/.ssh/authorized_keys" ]  && send_to_s3 "${HOME}/.ssh/authorized_keys" $1
