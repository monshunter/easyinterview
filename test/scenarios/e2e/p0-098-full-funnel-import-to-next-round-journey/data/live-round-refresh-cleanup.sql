begin;

delete from ai_task_runs
where user_id = '019f6098-0000-7000-8000-000000000001'
   or resource_id in (
     select id from feedback_reports
     where user_id = '019f6098-0000-7000-8000-000000000001'
   )
   or resource_id in (
     select id from practice_sessions
     where user_id = '019f6098-0000-7000-8000-000000000001'
   );

delete from async_jobs
where (
    resource_type = 'feedback_report'
    and resource_id in (
      select id from feedback_reports
      where user_id = '019f6098-0000-7000-8000-000000000001'
    )
  )
  or (
    resource_type = 'auth_challenge'
    and resource_id in (
      select id from auth_challenges
      where email = 'p0-098-live-round-refresh@example.test'
    )
  );

delete from outbox_events
where aggregate_id in (
  select id from practice_sessions
  where user_id = '019f6098-0000-7000-8000-000000000001'
);

delete from audit_events
where user_id = '019f6098-0000-7000-8000-000000000001'
   or actor_id = '019f6098-0000-7000-8000-000000000001'
   or resource_id in (
     select id from auth_challenges
     where email = 'p0-098-live-round-refresh@example.test'
   )
   or resource_id in (
     select id from practice_sessions
     where user_id = '019f6098-0000-7000-8000-000000000001'
   )
   or resource_id in (
     select id from practice_plans
     where user_id = '019f6098-0000-7000-8000-000000000001'
   );

delete from auth_challenges
where email = 'p0-098-live-round-refresh@example.test';

delete from users
where id = '019f6098-0000-7000-8000-000000000001'
   or email = 'p0-098-live-round-refresh@example.test';

commit;
