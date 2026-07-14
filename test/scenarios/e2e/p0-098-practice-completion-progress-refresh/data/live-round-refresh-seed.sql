begin;

insert into users (
  id, email, display_name, status, profile_completed_at, terms_accepted_at,
  created_at, updated_at
) values (
  '019f6098-0000-7000-8000-000000000001',
  'p0-098-live-round-refresh@example.test',
  'P0.098 Round Refresh',
  'active',
  now(),
  now(),
  now(),
  now()
);

insert into user_settings (
  user_id, ui_language, preferred_practice_language, timezone,
  analytics_opt_in, created_at, updated_at
) values (
  '019f6098-0000-7000-8000-000000000001',
  'zh-CN',
  'zh-CN',
  'Asia/Shanghai',
  false,
  now(),
  now()
);

insert into resumes (
  id, user_id, title, display_name, language, parse_status, parsed_summary,
  raw_text, source_type, original_text, parsed_text_snapshot,
  structured_profile, created_at, updated_at
) values (
  '019f6098-0000-7000-8000-000000000002',
  '019f6098-0000-7000-8000-000000000001',
  'P0.098 Platform Resume',
  'P0.098 Platform Resume',
  'zh-CN',
  'ready',
  '{}'::jsonb,
  'Go platform engineer with Kubernetes, GitOps, CI/CD and AI workflow experience.',
  'paste',
  'Go platform engineer with Kubernetes, GitOps, CI/CD and AI workflow experience.',
  'Go platform engineer with Kubernetes, GitOps, CI/CD and AI workflow experience.',
  '{"summary":"Go platform engineer","skills":["Go","Kubernetes","GitOps"]}'::jsonb,
  now(),
  now()
);

insert into target_jobs (
  id, user_id, resume_id, status, analysis_status, title, company_name,
  location_text, target_language, raw_jd_text, summary,
  fit_summary, created_at, updated_at
) values (
  '019f6098-0000-7000-8000-000000000003',
  '019f6098-0000-7000-8000-000000000001',
  '019f6098-0000-7000-8000-000000000002',
  'draft',
  'ready',
  'Platform Engineer',
  'P0.098 Systems',
  'Guangzhou',
  'zh-CN',
  'Build and operate a Go, Kubernetes and GitOps platform.',
  '{
    "coreThemes":["platform engineering","Kubernetes","GitOps"],
    "interviewRounds":[
      {"sequence":1,"type":"hr","name":"HR","durationMinutes":30,"focus":"motivation and role fit"},
      {"sequence":2,"type":"technical","name":"Technical","durationMinutes":30,"focus":"Go and Kubernetes system design"},
      {"sequence":4,"type":"manager","name":"Manager","durationMinutes":45,"focus":"ownership and delivery"}
    ],
    "provenance":{
      "promptVersion":"v0.1.0","rubricVersion":"v0.1.0",
      "modelId":"scenario-seed","language":"zh-CN",
      "featureFlag":"none","dataSourceVersion":"target-job.v1"
    }
  }'::jsonb,
  '{
    "strengths":["platform engineering"],"gaps":[],"riskSignals":[],
    "provenance":{
      "promptVersion":"v0.1.0","rubricVersion":"v0.1.0",
      "modelId":"scenario-seed","language":"zh-CN",
      "featureFlag":"none","dataSourceVersion":"target-job.v1"
    }
  }'::jsonb,
  now(),
  now()
);

insert into practice_plans (
  id, user_id, target_job_id, goal, interviewer_persona, difficulty,
  language, time_budget_minutes, resume_id, status, round_id,
  round_sequence, created_at, updated_at
) values (
  '019f6098-0000-7000-8000-000000000010',
  '019f6098-0000-7000-8000-000000000001',
  '019f6098-0000-7000-8000-000000000003',
  'baseline',
  'hiring_manager',
  'standard',
  'zh-CN',
  30,
  '019f6098-0000-7000-8000-000000000002',
  'ready',
  'round-1-hr',
  1,
  now(),
  now()
);

insert into practice_sessions (
  id, user_id, plan_id, target_job_id, status, language, started_at,
  created_at, updated_at
) values (
  '019f6098-0000-7000-8000-000000000020',
  '019f6098-0000-7000-8000-000000000001',
  '019f6098-0000-7000-8000-000000000010',
  '019f6098-0000-7000-8000-000000000003',
  'waiting_user_input',
  'zh-CN',
  now(),
  now(),
  now()
);

insert into practice_session_events (
  id, session_id, seq_no, event_type, payload, created_at
) values (
  '019f6098-0000-7000-8000-000000000021',
  '019f6098-0000-7000-8000-000000000020',
  1,
  'session_started',
  '{"source":"E2E.P0.098"}'::jsonb,
  now()
);

insert into practice_messages (
  id, session_id, seq_no, role, content, created_at
) values (
  '019f6098-0000-7000-8000-000000000022',
  '019f6098-0000-7000-8000-000000000020',
  1,
  'assistant',
  '请介绍一次你主导的平台工程改进。',
  now()
);

insert into practice_messages (
  id, session_id, seq_no, role, content, client_message_id,
  reply_status, reply_generation, created_at
) values (
  '019f6098-0000-7000-8000-000000000023',
  '019f6098-0000-7000-8000-000000000020',
  2,
  'user',
  '我主导了 Kubernetes 发布链路的 GitOps 改造。',
  '019f6098-0000-7000-8000-000000000029',
  'complete',
  1,
  now()
);

insert into practice_messages (
  id, session_id, seq_no, role, content, reply_to_message_id, created_at
) values (
  '019f6098-0000-7000-8000-000000000024',
  '019f6098-0000-7000-8000-000000000020',
  3,
  'assistant',
  '你如何衡量这次改造的效果？',
  '019f6098-0000-7000-8000-000000000023',
  now()
);

commit;
