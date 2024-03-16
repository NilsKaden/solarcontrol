# solarcontrol
Solarcontrol ensures battery safety when running grid tie inverters off a battery 
loaded by Victron Smart Solar MPPTs.

Currently only logs BatteryVoltage etc. from Victron MPPT. 

Uses Victron instant readout to get 
- Battery current in Amps
- Battery voltage in Volts
- Todays yield in kWh
- Current PV power in W

Some time in the future this code will be used to shut off any AhoyDTU compatible inverter
if the voltage of a battery connected to a Victron Smart Solar MPPT gets too low.
Additionally, we can turn on the inverter, or increase power, if the battery gets close to being full.


## How to run

Make sure to set the environment variables VICTRON_UUID and VICTRON_KEY. 
I recommend filling in the values in the .envrc-sample and renaming it to .envrc 

Can be comfortably loaded with [direnv](https://direnv.net/).

VICTRON_UUID can be found using a bluetooth scanner for your system. nRF Connect from Nordic works great.
[How to find the Victron instant readout encryption key](https://community.victronenergy.com/questions/187303/victron-bluetooth-advertising-protocol.html)

InverterInfo Response when shut down: 
{
    ID:0, 
    Enabled:true, 
    Name:\"HM-300\", 
    Serial:\"REMOVED\", 
    Version:\"0\", 
    PowerLimitRead:65535, 
    PowerLimitAck:false, 
    MaxPwr:0, TsLastSuccess:0, 
    Generation:1, 
    Status:0, 
    AlarmCnt:0,
    Rssi:0, 
    TsMaxAcPwr:0, 
    Ch:[][]int{[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, []int{0, 0, 0, 0, 0, 0, 0}},
    ChName:[]string{\"AC\", \"\"},
    ChMaxPwr:[]interface {}{interface {}(nil), 400}
}

InverterInfo Response when powered on:
{
    ID:0, 
    Enabled:true, 
    Name:\"HM-300\", 
    Serial:\"112182848743\", 
    Version:\"10014\",
    PowerLimitRead:36, 
    PowerLimitAck:false, 
    MaxPwr:300, 
    TsLastSuccess:1710550921, 
    Generation:1, 
    Status:2, 
    AlarmCnt:1, 
    Rssi:-75, 
    TsMaxAcPwr:1710550741, 
    Ch:[][]float32{[]float32{232.9, 0.48, 112, 50, 1, 22.1, 5.134, 7, 117.3, 95.482, 0.1, 112}, []float32{24.8, 4.76, 117.3, 7, 5.134, 29.325, 117.3}}, 
    ChName:[]string{\"AC\", \"\"}, 
    ChMaxPwr:[]interface {}{interface {}(nil), 400}}"
}

PowerLimitRead is always in %, it seems
