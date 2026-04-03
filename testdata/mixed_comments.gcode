; file header - mixed comment styles
(setup begin)
G28 ; home all axes
G90 (absolute positioning)
M82 ; absolute extrusion
G92 E0 (reset extruder)
; movement section
G1 Z5 F3000 ; lift nozzle
G1 X50 Y50 F6000 (rapid move to center)
G1 Z0.2 F1000 ; lower to print height
(end of setup)
; begin printing
G1 X100 Y50 E5 F1200
