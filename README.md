# Rancher Access Migrator

`rancher-access-migrator` is a distinct tool derived from Rancher Labs' `cattle-drive`. It provides safe logical export/import for same-Rancher and cross-Rancher project/RBAC migration. The logical planner is unit-tested; live Rancher integration is not claimed in this workspace.

A tool to migrate Rancher objects created for downstream cluster from a source to a target cluster, these objects include, but not limited to:

- Projects
  - Namespaces
  - ProjectRoleTemplateBindings
- ClusterRoleTemplateBindings
- Cluster Apps
- Cluster Catalog Repos

The portable workflow migrates Projects, namespace assignments, PRTBs, CRTBs, and referenced custom RoleTemplates. It does not migrate workloads, Secrets, auth-provider configuration, global permissions, or credentials.

## Portable export/import

```sh
rancher-access-migrator export --cluster old --output migration.yaml --kubeconfig rancher-a.yaml
rancher-access-migrator validate --file migration.yaml
rancher-access-migrator import --file migration.yaml --target-cluster new --kubeconfig rancher-b.yaml --dry-run
rancher-access-migrator import --file migration.yaml --target-cluster new --kubeconfig rancher-b.yaml
```

The snapshot stores logical project names and principals, not destination IDs. Import inventories destination users/principals and RoleTemplates, reports create/remap/skip/unresolved actions, and blocks unresolved dependencies unless `--allow-unresolved` is explicitly supplied. Binding names are deterministic, and projects/roles are matched by logical name for idempotent reruns.

## Cross-Rancher direct migration

Upstream already includes an experimental second-kubeconfig path:

```sh
rancher-access-migrator status -s old -t new --kubeconfig rancher-a.yaml --target-rancher-config rancher-b.yaml
rancher-access-migrator migrate -s old -t new --kubeconfig rancher-a.yaml --target-rancher-config rancher-b.yaml
```

Tokens should remain inside protected kubeconfig files or `KUBECONFIG`; they are not printed. No Rancher compatibility matrix is claimed until the integration plan in `docs/INTEGRATION_TESTING.md` has been run for each intended version.

See `docs/ARCHITECTURE.md`, `docs/MIGRATION_WORKFLOW.md`, `docs/VALIDATION.md`, and `docs/ROLLBACK.md`.

## Usage

First you would need a kubeconfig that can connect to the local cluster of the Rancher environment with admin access, for more information on how to obtain this please visit the [docs](https://ranchermanager.docs.rancher.com/api/quickstart), the tool has 3 subcommands:

### Status

The status subcommand will list all the related objects and their status, the status can be one of three:

- Migrated
- Not Migrated
- Migrated but with wrong spec

```sh
$ rancher-access-migrator status -s hussein-rke1 -t hgalal-rke2 --kubeconfig kubeconfig.yaml
Project status:
 - [test-project] ✔
  -> users permissions:
	 - [prtb-kds2g] ✔
  -> namespaces:
Cluster users permissions:
 - [crtb-p7cpc] ✔
 - [crtb-v9ls4] ✔
Catalog repos:
 - [k3k] ✔
```

### Migrate

The migrate subcommand will migrate all related objects to to the target downstream cluster, note that the some objects are only created on the local cluster while some objects has to be created on the downstream cluster itself.

```sh
$ rancher-access-migrator migrate -s hussein-rke1 -t hgalal-rke2 --kubeconfig kubeconfig.yaml
Migrating Objects from cluster [hussein-rke1] to cluster [hgalal-rke2]:
- migrating Project [migrate-project]... Done.
```

### Interactive

The interactive subcommands allows you to navigate in a simple list menu through all the objects and their status, and allows you to migrate certain object individually.

[![asciicast](https://asciinema.org/a/Bd6wc7pT0RM92sWqOctAanReL.svg)](https://asciinema.org/a/Bd6wc7pT0RM92sWqOctAanReL)
