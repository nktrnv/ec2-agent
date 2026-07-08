#!/bin/sh
set -e

systemctl stop ec2-agent.timer ec2-agent.service 2>/dev/null || true
systemctl disable ec2-agent.timer 2>/dev/null || true
systemctl daemon-reload
