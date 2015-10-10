# slackem

Send a slack message from the command line.

## Installation

    $ go get github.com/parkr/slackem

## Configuration

Mandatory:

    $ export SLACK_WEBHOOK_URL="http://slack.com/my/incoming/webhook/"

Optional:

    $ export SLACK_USERNAME="herro"
    $ export SLACK_ICON_EMOJI=":lips:"

## Usage

    $ slackem general Hey everyone, please rate the app.
    $ slackem -color=red general Hey everyone, please rate the app.

Colors available: grey, green, red, blue

## Wishlist

1. Take input from STDIN (except channel & color)
2. Better arg parsing with options
