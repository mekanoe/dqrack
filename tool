#!/bin/bash

read -r -d '' help <<HELP
💖      dqrack      💞
💘 Swiss Army Knife 💓
💝     ~~*~~*~~     💕

*** 💁 Available Commands ->

install
  dep ensures, generates, then builds the project.

env [set|unset|dot] [path]
  set relevant environment variables
  * use like \`eval \$(./tool env)\`
  * ex: eval \$(./tool env)
  * ex: eval \$(./tool env unset)
  * for dot, if path isn't set, i write to stdout.

ui service
  try to open a service's web ui
  * usable for: dgraph
  * ex: ./tool ui dgraph

help
  this message
HELP

OPEN="$(uname | grep -q Linux && echo "xdg-")open"

cmd_install() {
    echo "🚚 updating dependencies"
    dep ensure # currently a bad idea.

    echo "⚙️ running generate tasks"
    go generate ./...

    echo "💫 installing"
    go install ./...
}

cmd_env_set() {
    SET=export
    SPLIT="="

    if [[ "$SHELL" == *fish* ]]; then
        SET="set -gx"
        SPLIT=" "
    fi

    echo "$SET DGRAPH_ADDR$SPLIT$(docker-compose port dgraph 9080);"
    
    (>&2 echo "✅ Environment set.")    
}

cmd_env_dot() {
    echo "DGRAPH_ADDR=$(docker-compose port dgraph 9080)" >> $1
    
    (>&2 echo "📝  .env file written")    
}

cmd_env_unset() {
    SET=export
    SPLIT="="

    if [[ "$SHELL" == *fish* ]]; then
        SET="set -e"
        SPLIT=" "
    fi

    echo "$SET DGRAPH_ADDR;"

    (>&2 echo "❌ Environment cleared.")
}

cmd_ui() {
    url="http://$(docker-compose port $1 8080)"
    echo "🛠 trying to open $url"
    $OPEN $url
}

main() {
    case "$1" in
        help) echo "$help";;
        env) cmd_env_${2:-set} "${3:-/dev/stdout}";;
        ui) cmd_ui $2;;
        install) cmd_install;;
        *) echo "$help";;
    esac
}

main "$@"
