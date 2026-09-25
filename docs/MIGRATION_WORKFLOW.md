# Migration workflow

1. Take a supported Rancher management backup.
2. Export while the source registration exists; protect the snapshot as access metadata.
3. Import/register the target cluster and configure the equivalent authentication provider.
4. Run `validate`, then `import --dry-run`.
5. Resolve every missing principal, role, duplicate project name, and namespace conflict.
6. Apply and rerun dry-run/status for verification.

Do not run agents from two Rancher deployments against one downstream cluster simultaneously unless the applicable Rancher procedure explicitly permits it.
