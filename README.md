# TFPrettyPlan

TFPrettyPlan is a command-line tool that renders Terraform plan files in a more human-readable format. It parses Terraform plan JSON files and displays the changes in a nicely formatted ASCII table, highlighting resources that will be created, updated, or deleted.

## Features

- Parses Terraform plan JSON files
- Displays a summary of resources to be created, updated, or deleted
- Shows detailed information about each resource change
- Highlights changes with color (can be disabled)
- Supports reading from files or standard input

## Installation

### Using Go

```bash
go install github.com/ao/tfprettyplan/cmd/tfprettyplan@latest
```

### From Source

```bash
git clone https://github.com/ao/tfprettyplan.git
cd tfprettyplan
go build -o tfprettyplan ./cmd/tfprettyplan
```

## Usage

TFPrettyPlan can read Terraform plan files in JSON format. You can provide the plan file as an argument or pipe the JSON data to the tool.

### Basic Usage

```bash
# Read from a file
tfprettyplan plan.json

# Using the -file flag
tfprettyplan -file plan.json

# Pipe from terraform show
terraform show -json plan.tfplan | tfprettyplan
```

### Flags

- `-file, -f`: Path to Terraform plan JSON file
- `-no-color`: Disable color output
- `-version, -v`: Show version information

## Example

To use TFPrettyPlan with a Terraform plan:

1. Generate a Terraform plan file:
   ```bash
   terraform plan -out=plan.tfplan
   ```

2. Convert the plan to JSON:
   ```bash
   terraform show -json plan.tfplan > plan.json
   ```

3. Use TFPrettyPlan to visualize the plan:
   ```bash
   tfprettyplan plan.json
   ```

   Or in a single command:
   ```bash
   terraform show -json plan.tfplan | tfprettyplan
   ```

## Sample Output

```
Terraform Plan Summary
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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
