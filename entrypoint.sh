#!/bin/sh

# set safe.directory to the current working directory
sh -c "git config --global --add safe.directory $PWD"

# pass appId flag from GH_APP_ID environment variable if it exists
if [ -n "$GH_APP_ID" ]; then
  set -- "$@" -i "$GH_APP_ID"
fi

# pass pem file flag from GH_APP_PRIVATE_KEY_FILE environment variable if it exists
if [ -n "$GH_APP_PRIVATE_KEY_FILE" ]; then
  set -- "$@" -p "$GH_APP_PRIVATE_KEY_FILE"
fi

# pass repository flag from REPOSITORY environment variable if it exists
if [ -n "$REPOSITORY" ]; then
  set -- "$@" -r "$REPOSITORY"
fi

# pass branch flag from BRANCH environment variable if it exists
if [ -n "$BRANCH" ]; then
  set -- "$@" -b "$BRANCH"
fi

# pass HEAD branch flag from HEAD_BRANCH environment variable if it exists
if [ -n "$HEAD_BRANCH" ]; then
  set -- "$@" -h "$HEAD_BRANCH"
fi

# pass force flag from FORCE_PUSH environment variable if it exists
if [ -n "$FORCE_PUSH" ]; then
  set -- "$@" -f
fi

# pass tags flag from TAGS environment variable if it exists
if [ -n "$TAGS" ]; then
  set -- "$@" -t "$TAGS"
fi

# pass message flag from COMMIT_MSG environment variable if it exists
if [ -n "$COMMIT_MSG" ]; then
  set -- "$@" -m "$COMMIT_MSG"
fi

# pass addNewFiles flag ADD_NEW_FILES environment variable if it exists
if [ -n "$ADD_NEW_FILES" ]; then
  set -- "$@" -a
fi

# pass coauthors flag from COAUTHORS environment variable if it exists
if [ -n "$COAUTHORS" ]; then
  set -- "$@" -c "$COAUTHORS"
fi

# print the version of the action
echo "Github app commit action: $(/bin/action -version)\n"

# execute the action with the arguments
/bin/action "$@"