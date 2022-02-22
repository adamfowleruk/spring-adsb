CREATE TABLE tracks
(
  craftid varchar(128),
  gpsdatetime timestamp,
  gpsdate varchar(10),
  lon double precision,
  lat double precision,
  jsondata json
) DISTRIBUTED RANDOMLY;
