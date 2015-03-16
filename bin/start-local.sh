#!/bin/sh

make
supervisord -n -c supervisord-local.conf
