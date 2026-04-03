; Marlin print start sequence
G28 ; home all axes
M104 S200 ; set hotend temp
M140 S60 ; set bed temp
M109 S200 ; wait for hotend
M190 S60 ; wait for bed
G29 ; auto bed leveling
G92 E0 ; reset extruder
G1 Z5 F3000 ; lift nozzle
G1 X0.1 Y20 F5000 ; move to prime position
G1 Z0.3 F3000 ; lower nozzle
G1 X0.1 Y200 E15 F1500 ; prime line
G1 X0.4 Y200 E15 F5000 ; move over
G1 X0.4 Y20 E30 F1500 ; second prime line
G92 E0 ; reset extruder
G1 Z2 F3000 ; lift nozzle
