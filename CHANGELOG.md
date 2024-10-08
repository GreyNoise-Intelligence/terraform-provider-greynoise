
## v0.6.0 - 2024-09-20
### FEATURES
* resource/greynoise_sensor_bootstrap: Parses `public_ip` as a comma-separated list of IPs or CIDRs into `sensor_public_ips`.

## v0.5.0 - 2024-09-18
### FEATURES
* * resource/greynoise_sensor_metadata: Add resource to manage sensor metadata

## v0.4.1 - 2024-09-13
### BUG FIXES
* resource/greynoise_sensor_bootstrap: Fixes bug with changing value of computed SSH port selected.

## v0.4.0 - 2024-09-13
### ENHANCEMENTS
* resource/greynoise_sensor_bootstrap: Add `nat` argument to specify is NAT is used for traffic to bootstrap server.
* resource/greynoise_sensor_bootstrap: Adds `config` argument to allow specifying values for use in provisioners.

## v0.3.0 - 2024-09-11
### ENHANCEMENTS
* resource/greynoise_sensor_bootstrap: Adds unbootstrap script to resource for destroy

## v0.2.0 - 2024-09-06
### ENHANCEMENTS
* data-source/greynoise_sensor: Additional sensor properties added to computed values.
### NOTES
* Provider documentation updated to remove typo in previous release.

## v0.1.0 - 2024-09-04
### FEATURES
* provider: Added provider implementation
* data-source/greynoise_personas: Data source to lookup GreyNoise personas based on a set of filter criteria.
* data-source/greynoise_account: Data source to lookup current account information.
* data-source/greynoise_sensor: Data source to lookup a sensor based on public IP. 
* resource/greynoise_sensor_bootstrap: Resource to provide sensor bootstrapping scripts and contain remote-exec installtion of sensor. 
* resource/greynoise_sensor_persona: Resource to manage sensor personas.
### NOTES
* Documentation for GreyNoise provider
