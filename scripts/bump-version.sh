NEW_APP_VERSION=$1

if [ -z "$NEW_APP_VERSION" ]; then
  echo "NEW_APP_VERSION is not set"
  exit 1
fi

echo "Bumping the app version to $NEW_APP_VERSION"
sed -i '' "s/Version =.*/Version = \"$NEW_APP_VERSION\"/" main.go