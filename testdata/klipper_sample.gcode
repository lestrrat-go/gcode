; synthetic Klipper-flavoured slicer output covering the
; extended-command shapes the streaming Reader has to handle
; (named args, lists, quoted strings, mixed case, long lines).

; ---- preamble ----
M73 P0 R10
M140 S60
M104 S200
M109 S200
M190 S60

; ---- Klipper extended commands ----
EXCLUDE_OBJECT_DEFINE NAME=part_0 CENTER=120,120 POLYGON=[[100,100],[140,100],[140,140],[100,140]]
EXCLUDE_OBJECT_DEFINE NAME=part_1 CENTER=80,80 POLYGON=[[60,60],[100,60],[100,100],[60,100]]
EXCLUDE_OBJECT_START NAME=part_0
SET_PRINT_STATS_INFO TOTAL_LAYER=10 CURRENT_LAYER=1
SET_PRESSURE_ADVANCE ADVANCE=0.04 SMOOTH_TIME=0.04
SET_FAN_SPEED FAN=cooling SPEED=0.5
SET_VELOCITY_LIMIT VELOCITY=200 ACCEL=3000 SQUARE_CORNER_VELOCITY=5
BED_MESH_PROFILE LOAD=default
SAVE_GCODE_STATE NAME=before_purge
RESTORE_GCODE_STATE NAME=before_purge MOVE=1 MOVE_SPEED=60
TIMELAPSE_TAKE_FRAME

; case is preserved on extended-arg keys
BED_MESH_CALIBRATE mesh_min=112.4,114.0 mesh_max=161.8,209.2 ALGORITHM=bicubic PROBE_COUNT=4,4

; ---- classic moves with comments ----
G28 ; home all axes
G92 E0 ; reset extruder
G1 Z5 F3000 ; lift nozzle
G1 X0.1 Y20 F5000 ; move to prime
G1 X0.1 Y200 E15 F1500 ; prime line

; ---- mixed comment forms ----
(setup begin)
G90
M82
(end of setup)

; ---- line numbers + checksum ----
N1 G28*18
N2 G1 X10 Y10 E1*121

; ---- string-arg command (free-form tail; structured args are empty) ----
M117 Hello World
M118 // bridging next layer

EXCLUDE_OBJECT_END NAME=part_0
M73 P100 R0
