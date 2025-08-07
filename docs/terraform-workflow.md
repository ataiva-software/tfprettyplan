# Working with Terraform Plans in TFPrettyPlan

## Understanding the Error

The error you're seeing:
```
Error: Failed to load plugin schemas
...
Error parsing plan JSON: failed to parse JSON: unexpected end of JSON input
```

This happens because when you run `terraform show -json` on a plan file, Terraform needs access to all provider plugins that were used to create the plan. If those aren't available in your current environment, Terraform can't generate the JSON output.

## Correct Workflow for Using TFPrettyPlan

To properly use TFPrettyPlan with Terraform plans, follow these steps:

### Method 1: Generate JSON in the Same Directory (Recommended)

1. **Navigate to your Terraform project directory**:
   ```bash
   cd /path/to/your/terraform/project
   ```

2. **Initialize Terraform** (if not already done):
   ```bash
   terraform init
   ```

3. **Create the plan and generate JSON in one step**:
   ```bash
   terraform plan -out=plan.tfplan && terraform show -json plan.tfplan > plan.json
   ```

4. **Use TFPrettyPlan with the JSON file**:
   ```bash
   tfprettyplan plan.json
   ```

### Method 2: Direct Pipe (Same Directory)

If you want to skip saving the JSON file:

```bash
terraform plan -out=plan.tfplan && terraform show -json plan.tfplan | tfprettyplan
```

## Troubleshooting Provider Errors

If you still encounter provider schema errors:

1. **Ensure providers are installed**:
   ```bash
   terraform init
   ```

2. **Check provider versions**:
   Ensure the providers in your current environment match those used to create the plan.

3. **Use a saved JSON file**:
   Generate the JSON file in the environment where the plan was created, then copy the JSON file to use with TFPrettyPlan elsewhere.

## Using TFPrettyPlan with Existing Plan Files

If you have an existing `.tfplan` file from another directory:

1. **Copy the plan file to the original Terraform project directory**
2. **Run the commands there**:
   ```bash
   cd /original/terraform/project
   terraform show -json /path/to/plan.tfplan > plan.json
   ```

3. **Use the generated JSON file with TFPrettyPlan**:
   ```bash
   tfprettyplan plan.json
   ```

## Best Practices

- Always generate the JSON in the same directory where the Terraform configuration exists
- Keep the `.terraform` directory intact when generating JSON from plan files
- If sharing plans across environments, share the JSON output rather than the `.tfplan` file
