# Architecture and upstream analysis

The upstream CLI uses `urfave/cli`; Wrangler/Lasso clients access Rancher management API resources, while a typed core client updates downstream namespaces. `Cluster.Populate` loads non-default Projects, PRTBs, CRTBs, namespace mappings, catalog repositories, users, and global role bindings. `Compare` matches projects by display name; `Migrate` creates destination projects and rewrites dependent references to their new IDs.

Upstream already had undocumented experimental cross-Rancher support through `--target-rancher-config`, which builds an independent destination management client and downstream proxy client. Its gaps were portable capture, RoleTemplate migration, durable validation, safe principal resolution, and tests.

The new `pkg/migration` layer stores logical project names and principals separately from Rancher-generated IDs. Import order is validate, RoleTemplates, Projects, PRTBs/CRTBs, then namespace metadata. It never deletes source objects or workloads.
