#!/bin/sh

make -s
supervisord -n -c supervisord-local.conf
