# Installation

## Tech2 PassThru support
* Go to "C:\Program Files (x86)\GM\Tech2 J2534 DLL"
* Right click "Tech2_Config.exe" and "Run as administrator"
* Select one number higher COM port than your USB to Serial adapter is assigned to ( There is a bug in the Tech2 Config program )

### Cofigure USB to serial port adapter

* Start Device Manager and open Properties
* Under "Port Settings" press Advanced
* Under "BM Options" set Latency Timer to 2 (msec) ( raise this number gradually if you are experiencing connectivity problems )