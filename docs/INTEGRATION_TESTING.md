# Integration-test strategy

Use two disposable Rancher installations and disposable imported clusters. Seed custom projects, namespace mappings, user/group PRTBs, CRTBs, and a custom RoleTemplate. Exercise same-Rancher direct migration, cross-Rancher direct migration, export/import, missing-principal blocking, dry-run immutability, and a second idempotent run. Compare logical tuples and confirm workloads and Secrets are unchanged. Repeat for every Rancher release line intended for support. These live tests have not been run here.
