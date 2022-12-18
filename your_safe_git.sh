#!/bin/sh

# this is a 'safe' version of your_git.sh, which only runs mygit inside directory paths containing 'mygit-test'
# you can register it as an alias: 
# alias mygit="/path/to/your/repo/your_safe_git.sh"

set -e

case $(pwd) in 
	*mygit-test*) ;;
	*) echo "Error: Run this command in a directory path containing 'mygit-test'." && exit 0;;
esac

tmpFile=$(mktemp)

( cd $(dirname "$0") &&
	go build -o "$tmpFile" ./cmd/mygit )

exec "$tmpFile" "$@"
