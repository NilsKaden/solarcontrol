# solarcontrol
Solarcontrol ensures battery safety when running grid tie inverters off a battery charged by Victron Smart Solar MPPTs.

Turns off any AHOY DTU compatible inverter if the battery voltage gets too low.
Battery measurements are taken from victron MPPTs via bluetooth instant readout.
If the converter is connected to grid via a mystrom plug, we can turn that off, too.


## How to run

Make sure to set the environment variables 
- VICTRON_UUID 
- VICTRON_KEY
- AHOY_ENDPOINT
- INVERTER_ID
- SHUTOFF_VOLTAGE
- MYSTROM_ENDPOINT

I recommend filling in the values in the .envrc-sample and renaming it to .envrc 

Can be comfortably loaded with [direnv](https://direnv.net/).

VICTRON_UUID can be found using a bluetooth scanner for your system. nRF Connect from Nordic works great.
[How to find the Victron instant readout encryption key](https://community.victronenergy.com/questions/187303/victron-bluetooth-advertising-protocol.html)
