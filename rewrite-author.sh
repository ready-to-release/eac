#!/bin/bash
# Get current author email
EMAIL=$(git log -1 --format='%ae' HEAD)

case "$EMAIL" in
    test@example.com|casper.nielsen@gmail.com|casper.nielsen@nexigroup.com|user@local|*casperease*)
        git config user.name "Casper"
        git config user.email "casper@eac.dev"
        ;;
    *dependabot*)
        git config user.name "dependabot"
        git config user.email "dependabot@eac.dev"
        ;;
    miohansen@gmail.com|*miohansen@users.noreply*)
        git config user.name "Michael"
        git config user.email "michael@eac.dev"
        ;;
    code@tomasmalmsten.com)
        git config user.name "Tomas"
        git config user.email "tomas@eac.dev"
        ;;
esac

git commit --amend --reset-author --no-edit
