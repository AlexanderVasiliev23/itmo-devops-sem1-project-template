#!/bin/bash

# подгружаем переменные окружения из .env файла
if [ -f .env ]; then
    export $(cat .env | xargs)
else
    echo "Файл с переменными окружения .env не найден. Он должен находиться в корне проекта"
    exit 1
fi

go run ./cmd/. > logs.log 2>&1 &