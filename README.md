fank/steamcmd-gmail
===================

Steamcmd with Steam Guard support by using GMail API

# Overview

This Docker container is used for automated updates with steamcmd for people who want to use there steam account which has Steam Guard enabled, and don't want to turn it off.

# What does it?

When the container starts it will start a daemon process which watches steamcmd and if you need to enter the steam guard code, a GMail API call will get your latest mail and search for the Steam Guard code, if it was found it will be send to steamcmd to authentificate.

# Installation

## Initial
For GMail you need to authentificate your account a credential file will be saved this should be stored secure:

`docker run -t --rm -v ~/.gmail-credential.json:/credential.json fank/steamcmd-gmail +quit`

If you don't trust me you can use our own client secret file, follow "Step 1" https://developers.google.com/gmail/api/quickstart/go

And mount your `client_secret.json`:

`docker run -t --rm -v <your client_secret.json>:/client_secret.json -v ~/.gmail-credential.json:/credential.json  fank/steamcmd-gmail +quit`

## Automated

e.g. Starbound server:

`docker run -t --rm -v ~/.gmail-credential.json:/credential.json -v <your server folder>:/server fank/steamcmd-gmail +login <username> <password> +force_install_dir /server +app_update 211820 validate +quit`
