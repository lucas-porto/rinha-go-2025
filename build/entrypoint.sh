#!/bin/sh

if [ -f "./api" ]; then
    exec "./api"
elif [ -f "./worker" ]; then
    exec "./worker"
else
    echo "Erro: Nenhum binário encontrado"
    ls -la
    exit 1
fi 