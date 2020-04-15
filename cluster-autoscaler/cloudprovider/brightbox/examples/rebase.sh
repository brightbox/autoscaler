#!/bin/sh

set -e

git rebase --onto cluster-autoscaler-1.17.2 cluster-autoscaler-1.17.1 autoscaler-brightbox-cloudprovider-1.17
git rebase --onto cluster-autoscaler-1.18.1 cluster-autoscaler-1.18.0 autoscaler-brightbox-cloudprovider-1.18

