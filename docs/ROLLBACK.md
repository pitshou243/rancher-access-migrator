# Rollback and recovery

The tool never deletes source state. Keep the export and a supported Rancher backup. On failure, correct the reported dependency and rerun; deterministic resources make this preferable to destructive cleanup. Automatic rollback is deliberately omitted because it could remove pre-existing access. Manually remove only verified newly created resources, or restore the management backup when the complete Rancher deployment must be reverted.
