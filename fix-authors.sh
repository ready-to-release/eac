#!/bin/bash
case "$GIT_AUTHOR_EMAIL" in
    test@example.com|casper.nielsen@gmail.com|casper.nielsen@nexigroup.com|user@local)
        export GIT_AUTHOR_NAME="Casper"
        export GIT_AUTHOR_EMAIL="casper@eac.dev"
        ;;
    *casperease*)
        export GIT_AUTHOR_NAME="Casper"
        export GIT_AUTHOR_EMAIL="casper@eac.dev"
        ;;
    *dependabot*)
        export GIT_AUTHOR_NAME="dependabot"
        export GIT_AUTHOR_EMAIL="dependabot@eac.dev"
        ;;
    miohansen@gmail.com)
        export GIT_AUTHOR_NAME="Michael"
        export GIT_AUTHOR_EMAIL="michael@eac.dev"
        ;;
    *miohansen@users.noreply*)
        export GIT_AUTHOR_NAME="Michael"
        export GIT_AUTHOR_EMAIL="michael@eac.dev"
        ;;
    code@tomasmalmsten.com)
        export GIT_AUTHOR_NAME="Tomas"
        export GIT_AUTHOR_EMAIL="tomas@eac.dev"
        ;;
esac
export GIT_COMMITTER_NAME="$GIT_AUTHOR_NAME"
export GIT_COMMITTER_EMAIL="$GIT_AUTHOR_EMAIL"
