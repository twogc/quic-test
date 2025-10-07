#!/bin/bash

# Скрипт для обновления версии в файле tag.txt

if [ $# -eq 0 ]; then
    echo "Использование: $0 <версия>"
    echo "Пример: $0 v1.2.3"
    exit 1
fi

VERSION=$1

# Проверяем, что версия начинается с 'v'
if [[ ! $VERSION =~ ^v ]]; then
    echo "Версия должна начинаться с 'v' (например: v1.2.3)"
    exit 1
fi

# Обновляем файл tag.txt
echo "$VERSION" > tag.txt

echo "Версия обновлена до: $VERSION"
echo "Содержимое tag.txt:"
cat tag.txt

# Проверяем, что версия работает
echo ""
echo "Проверка версии:"
go build . && ./quic-test --version