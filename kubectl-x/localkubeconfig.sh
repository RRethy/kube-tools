#!/bin/bash

function klocal {
    export KUBECONFIG=$(kubectl-x kubeconfig copy)
}
