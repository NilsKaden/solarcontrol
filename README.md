# solarcontrol
Solarcontrol ensures battery safety when running grid tie inverters off a battery 
loaded by Victron Smart Solar MPPTs.

Currently only logs BatteryVoltage etc. from Victron MPPT. 

Uses Victron instant readout to get 
- Battery current in Amps
- Battery voltage in Volts
- Todays yield in kWh

Some time in the future this code will be used to shut off any AhoyDTU compatible inverter
if the voltage of a battery connected to a Victron Smart Solar MPPT gets too low.
Additionally, we can turn on the inverter, or increase power, if the battery gets close to being full.


## How to run

Make sure to set the environment variables VICTRON_UUID and VICTRON_KEY. 
I recommend filling in the values in the .envrc-sample and renaming it to .envrc 
Can be comfortably loaded with [direnv](https://direnv.net/).

VICTRON_UUID can be found using a bluetooth scanner for your system. nRF Connect from Nordic works great.
[How to find the Victron instant readout encryption key](https://community.victronenergy.com/questions/187303/victron-bluetooth-advertising-protocol.html)
