# Seed Input

- Valid URL: `https://jobs.example.com/role/1?token=secret#frag`.
- Invalid URL: `http://169.254.169.254/latest/meta-data`.
- Fetch result: sanitized public URL plus fixture JD body.
- AI: deterministic fake `target.import.default` response.
