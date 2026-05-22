ALTER TABLE daily_records ADD COLUMN exp_earned INTEGER NOT NULL DEFAULT 0;
UPDATE daily_records
   SET exp_earned = (SELECT exp_per_done FROM habits WHERE habits.id = daily_records.habit_id)
 WHERE exp_earned = 0;
