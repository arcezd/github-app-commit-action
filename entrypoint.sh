#!/bin/sh
args="$@"

# pass pem file flag from GITHUB_APP_PRIVATE_KEY_FILE environment variable if it exists
if [ -n "$GITHUB_APP_PRIVATE_KEY_FILE" ]; then
  args="$args -p $GITHUB_APP_PRIVATE_KEY_FILE"
fi

# pass repository flag from REPOSITORY environment variable if it exists
if [ -n "$REPOSITORY" ]; then
  args="$args -r $REPOSITORY"
fi

# pass branch flag from BRANCH environment variable if it exists
if [ -n "$BRANCH" ]; then
  args="$args -b $BRANCH"
fi

# pass HEAD branch flag from HEAD_BRANCH environment variable if it exists
if [ -n "$HEAD_BRANCH" ]; then
  args="$args -h $HEAD_BRANCH"
fi

# pass message flag from COMMIT_MSG environment variable if it exists
if [ -n "$COMMIT_MSG" ]; then
  args="$args -m $COMMIT_MSG"
fi

# pass addNewFiles flag ADD_NEW_FILES environment variable if it exists
if [ -n "$ADD_NEW_FILES" ]; then
  args="$args -a=$ADD_NEW_FILES"
fi

# pass coauthors flag from COAUTHORS environment variable if it exists
if [ -n "$COAUTHORS" ]; then
  args="$args -c $COAUTHORS"
fi

/bin/action $args