[![Build Status](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli.svg?branch=master)](https://travis-ci.org/pivotal-cf-experimental/diego-edge-cli)

diego-edge-cli
==============

Diego Edge CLI

Example Usage:
    
    DIEGO_EDGE_ETCD="http://192.168.11.11:4001" diego-edge-cli start Bingo-app -i "docker:///dajulia3/diego-edge-docker-app" -c "/dockerapp"
