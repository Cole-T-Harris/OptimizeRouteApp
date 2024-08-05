SELECT routes.id, routes.active, routes.user_id, routes.start_latitude, routes.start_longitude, routes.end_latitude, routes.end_longitude, routes.time_zone
FROM routes
WHERE routes.active = true;