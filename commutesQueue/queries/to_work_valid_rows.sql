SELECT
  routes.id,
  routes.user_id,
  routes.active,
  routes.start_latitude,
  routes.start_longitude,
  routes.end_latitude,
  routes.end_longitude,
  routes.time_zone
FROM
  routes
  INNER JOIN route_schedule ON routes.id = route_schedule.route_id
WHERE
  active = true
  AND CAST(NOW() AT TIME ZONE routes.time_zone AS TIME) 
      BETWEEN route_schedule.morning_start_time 
      AND route_schedule.morning_end_time
  AND (
    (EXTRACT(DOW FROM NOW() AT TIME ZONE routes.time_zone) = 0 AND route_schedule.sunday = true) OR
    (EXTRACT(DOW FROM NOW() AT TIME ZONE routes.time_zone) = 1 AND route_schedule.monday = true) OR
    (EXTRACT(DOW FROM NOW() AT TIME ZONE routes.time_zone) = 2 AND route_schedule.tuesday = true) OR
    (EXTRACT(DOW FROM NOW() AT TIME ZONE routes.time_zone) = 3 AND route_schedule.wednesday = true) OR
    (EXTRACT(DOW FROM NOW() AT TIME ZONE routes.time_zone) = 4 AND route_schedule.thursday = true) OR
    (EXTRACT(DOW FROM NOW() AT TIME ZONE routes.time_zone) = 5 AND route_schedule.friday = true) OR
    (EXTRACT(DOW FROM NOW() AT TIME ZONE routes.time_zone) = 6 AND route_schedule.saturday = true)
  );