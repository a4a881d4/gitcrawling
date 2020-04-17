#! /bin/bash

bin/mnp $1/packs $1/packn >> $1.log
bin/buildidx $1/packn >> $1.log
