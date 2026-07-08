#!/bin/sh
set -e

systemctl daemon-reload
systemctl enable --now ec2-agent.timer
