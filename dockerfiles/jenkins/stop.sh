#!/bin/bash

# ps -ef |grep java |awk '{print $2}'|xargs kill -9 
supervisorctl stop jenkins

