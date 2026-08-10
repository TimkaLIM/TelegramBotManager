ALTER TABLE schedules 
  ALTER COLUMN day_of_week TYPE VARCHAR(10) 
  USING CASE day_of_week::text
    WHEN '1' THEN 'Monday'
    WHEN '2' THEN 'Tuesday'
    WHEN '3' THEN 'Wednesday'
    WHEN '4' THEN 'Thursday'
    WHEN '5' THEN 'Friday'
    WHEN '6' THEN 'Saturday'
    WHEN '0' THEN 'Sunday'
    WHEN '7' THEN 'Sunday'
  END;