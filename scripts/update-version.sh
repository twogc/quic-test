#!/bin/bash
# Скрипт для обновления версии в tag.txt

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.2.3"
    exit 1
fi

VERSION=$1

# Проверяем формат версии
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ Invalid version format: $VERSION"
    echo "Expected format: v1.2.3"
    exit 1
fi

# Обновляем tag.txt
echo "$VERSION" > tag.txt

echo "✅ Version updated to $VERSION in tag.txt"
echo "📋 Next steps:"
echo "   1. Commit the change: git add tag.txt && git commit -m \"chore: bump version to $VERSION\""
echo "   2. Push to main: git push origin main"
echo "   3. GitHub Actions will automatically create tag and release"
