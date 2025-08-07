
+--------+-------+
| ACTION | COUNT |
+--------+-------+
| Create |     2 |
| Update |     1 |
| Delete |     1 |
| Total  |     4 |
+--------+-------+

Resources to Create
=====================

+--------+-------+
| ACTION | COUNT |
+--------+-------+
| Create |     2 |
| Update |     1 |
| Delete |     1 |
| Total  |     4 |
+--------+-------+

Resources to Create
==================

• aws_instance.example (aws_instance)

• aws_security_group.allow_ssh (aws_security_group)

Resources to Update
==================

• aws_s3_bucket.logs (aws_s3_bucket)
  +---------------+------------------+------------------+
  |   ATTRIBUTE   |    OLD VALUE     |    NEW VALUE     |
  +---------------+------------------+------------------+
  | acl           | private          | public-read      |
  | force_destroy | false            | true             |
  | tags.Name     | Log Bucket       | Logs Bucket      |
  +---------------+------------------+------------------+

Resources to Delete
==================

• aws_iam_role.lambda (aws_iam_role)
```

## Releases and Versioning

TFPrettyPlan uses semantic versioning and automated releases:

- New releases are automatically created when code is pushed to the main branch
- Releases follow semantic versioning (MAJOR.MINOR.PATCH)
- Each release includes pre-built binaries for all supported platforms
- Release notes are automatically generated from commit messages

You can find all releases on the [GitHub Releases page](https://github.com/ao/tfprettyplan/releases).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

For commit messages, we follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```
feat: add new feature
fix: fix a bug
docs: update documentation
refactor: code refactoring without functionality changes
```

This helps with automatic versioning and changelog generation.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
