import React, { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";

// Command templates for common GitHub Actions patterns
// These templates provide only steps to be inserted into existing jobs
const COMMAND_TEMPLATES = {
  "deploy-application": {
    name: "deploy-application",
    description:
      "Complete application deployment with GCP, database, and notifications",
    steps: `      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'
          cache: 'pip'

      - name: Setup Google Cloud CLI
        uses: google-github-actions/setup-gcloud@v2
        with:
          project_id: \${{ vars.PROJECT_ID }}

      - name: Authenticate to Google Cloud
        run: |
          gcloud auth activate-service-account --key-file=\${{ secrets.GOOGLE_APPLICATION_CREDENTIALS }}
          gcloud config set project \${{ vars.PROJECT_ID }}

      - name: Setup GCP resources
        id: gcp_setup
        timeout-minutes: 10
        run: |
          if ! gcloud compute instances describe \${{ vars.APP_NAME }}-vm --zone=\${{ vars.REGION }}-a >/dev/null 2>&1; then
            gcloud compute instances create \${{ vars.APP_NAME }}-vm \\
              --zone=\${{ vars.REGION }}-a \\
              --machine-type=e2-medium \\
              --image-family=debian-11 \\
              --image-project=debian-cloud \\
              --tags=webapp
          fi
          VM_IP=$(gcloud compute instances describe \${{ vars.APP_NAME }}-vm \\
            --zone=\${{ vars.REGION }}-a \\
            --format='get(networkInterfaces[0].accessConfigs[0].natIP)')
          echo "vm_external_ip=$VM_IP" >> $GITHUB_OUTPUT

      - name: Build application
        timeout-minutes: 10
        run: |
          python -m venv venv
          source venv/bin/activate
          if [ -f requirements.txt ]; then
            pip install -r requirements.txt
          fi
          if [ -f setup.py ]; then
            python setup.py build
            if [ -d tests ]; then
              python -m pytest tests/ || echo "Tests failed but continuing..."
            fi
            python setup.py sdist bdist_wheel
          fi

      - name: Deploy to VM
        timeout-minutes: 8
        run: |
          echo "\${{ secrets.SSH_PRIVATE_KEY }}" > ssh_key
          chmod 600 ssh_key
          ssh -i ssh_key -o StrictHostKeyChecking=no -o ConnectTimeout=10 \\
            deployment_user@\${{ steps.gcp_setup.outputs.vm_external_ip }} 'echo "SSH connection successful"'
          if [ -d dist ]; then
            scp -i ssh_key -o StrictHostKeyChecking=no -r dist/* \\
              deployment_user@\${{ steps.gcp_setup.outputs.vm_external_ip }}:/opt/webapp/
          fi
          rm -f ssh_key

      - name: Send deployment notification
        if: always()
        run: |
          STATUS="\${{ job.status }}"
          curl -X POST \\
            -H "Content-Type: application/json" \\
            -d "{\\"text\\": \\"Deployment $STATUS for \${{ vars.APP_NAME }} in \${{ vars.ENVIRONMENT }}\\"}" \\
            \${{ secrets.SLACK_WEBHOOK_URL }} || echo "Notification failed"`,
  },
  "gcloud-operations": {
    name: "gcloud-operations",
    description: "Google Cloud Platform resource management and operations",
    steps: `      - name: Setup Google Cloud CLI
        uses: google-github-actions/setup-gcloud@v2
        with:
          project_id: \${{ vars.PROJECT_ID }}

      - name: Authenticate to Google Cloud
        run: |
          gcloud auth activate-service-account --key-file=\${{ secrets.GOOGLE_APPLICATION_CREDENTIALS }}
          gcloud config set project \${{ vars.PROJECT_ID }}

      - name: Create VM instance
        run: |
          gcloud compute instances create \${{ vars.APP_NAME }}-vm \\
            --zone=us-central1-a \\
            --machine-type=e2-medium \\
            --image-family=debian-11 \\
            --image-project=debian-cloud \\
            --tags=webapp

      - name: Create storage bucket
        run: |
          gsutil mb gs://\${{ vars.PROJECT_ID }}-\${{ vars.APP_NAME }}-storage

      - name: List created resources
        run: |
          echo "=== Compute Instances ==="
          gcloud compute instances list
          echo "=== Storage Buckets ==="
          gsutil ls`,
  },
  "kubectl-deployment": {
    name: "kubectl-deployment",
    description: "Kubernetes cluster deployment and management",
    steps: `      - name: Setup kubectl
        uses: azure/setup-kubectl@v3
        with:
          version: 'latest'

      - name: Configure kubectl
        run: |
          gcloud container clusters get-credentials \${{ vars.CLUSTER_NAME }} \\
            --region=us-central1

      - name: Create namespace
        run: |
          kubectl create namespace \${{ vars.NAMESPACE }} \\
            --dry-run=client -o yaml | kubectl apply -f -

      - name: Deploy to Kubernetes
        run: |
          if [ -d k8s ]; then
            kubectl apply -f k8s/ -n \${{ vars.NAMESPACE }}
            kubectl rollout status deployment/\${{ vars.APP_NAME }} \\
              -n \${{ vars.NAMESPACE }} --timeout=300s
          else
            echo "No k8s directory found"
          fi

      - name: Get deployment status
        run: |
          kubectl get pods -n \${{ vars.NAMESPACE }}
          kubectl get services -n \${{ vars.NAMESPACE }}`,
  },
  "python-script": {
    name: "python-script",
    description: "Execute Python scripts with environment setup",
    steps: `      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'
          cache: 'pip'

      - name: Install dependencies
        run: |
          python -m venv venv
          source venv/bin/activate
          if [ -f requirements.txt ]; then
            pip install -r requirements.txt
          fi

      - name: Execute Python script
        run: |
          source venv/bin/activate
          python \${{ vars.SCRIPT_PATH }}

      - name: Run tests if available
        run: |
          source venv/bin/activate
          if [ -d tests ]; then
            python -m pytest tests/ -v
          fi`,
  },
  "git-operations": {
    name: "git-operations",
    description: "Git repository operations and management",
    steps: `      - name: Checkout with full history
        uses: actions/checkout@v4
        with:
          token: \${{ secrets.GITHUB_TOKEN }}
          fetch-depth: 0

      - name: Configure Git
        run: |
          git config --global user.name "GitHub Actions Bot"
          git config --global user.email "actions@github.com"

      - name: Create new branch
        run: |
          git checkout -b \${{ vars.BRANCH_NAME }}
          git push -u origin \${{ vars.BRANCH_NAME }}

      - name: Tag release
        run: |
          TAG_NAME="v$(date +%Y%m%d-%H%M%S)"
          git tag -a $TAG_NAME -m "Release $TAG_NAME"
          git push origin $TAG_NAME
          echo "Created tag: $TAG_NAME"`,
  },
  "ssh-deployment": {
    name: "ssh-deployment",
    description: "Deploy application to remote servers via SSH",
    steps: `      - name: Build application
        run: |
          if [ -f package.json ]; then
            npm ci && npm run build
          elif [ -f requirements.txt ]; then
            python -m pip install -r requirements.txt
            python setup.py build || echo "No setup.py found"
          fi

      - name: Deploy via SSH
        run: |
          echo "\${{ secrets.SSH_PRIVATE_KEY }}" > ssh_key
          chmod 600 ssh_key
          
          # Test SSH connection
          ssh -i ssh_key -o StrictHostKeyChecking=no -o ConnectTimeout=10 \\
            \${{ vars.SERVER_USER }}@\${{ vars.SERVER_HOST }} \\
            'echo "SSH connection successful"'
          
          # Create deployment directory
          ssh -i ssh_key -o StrictHostKeyChecking=no \\
            \${{ vars.SERVER_USER }}@\${{ vars.SERVER_HOST }} \\
            "sudo mkdir -p \${{ vars.DEPLOYMENT_PATH }}"
          
          # Transfer files
          if [ -d dist ]; then
            scp -i ssh_key -o StrictHostKeyChecking=no -r dist/* \\
              \${{ vars.SERVER_USER }}@\${{ vars.SERVER_HOST }}:\${{ vars.DEPLOYMENT_PATH }}/
          fi
          
          # Execute deployment commands
          ssh -i ssh_key -o StrictHostKeyChecking=no \\
            \${{ vars.SERVER_USER }}@\${{ vars.SERVER_HOST }} << 'EOF'
            cd \${{ vars.DEPLOYMENT_PATH }}
            if [ -f install.sh ]; then
              chmod +x install.sh && ./install.sh
            fi
            if systemctl list-unit-files | grep -q webapp; then
              sudo systemctl restart webapp
              sudo systemctl status webapp --no-pager
            fi
          EOF
          
          rm -f ssh_key`,
  },
  "api-webhook": {
    name: "api-webhook",
    description: "Call external APIs and webhooks",
    steps: `      - name: Call API endpoint
        run: |
          curl -X POST \\
            -H "Authorization: Bearer \${{ secrets.API_TOKEN }}" \\
            -H "Content-Type: application/json" \\
            -d '{"message": "Hello from GitHub Actions", "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}' \\
            \${{ vars.API_ENDPOINT }}

      - name: Send Slack notification
        if: always()
        run: |
          curl -X POST \\
            -H "Content-Type: application/json" \\
            -d '{
              "text": "API call completed to \${{ vars.API_ENDPOINT }}",
              "username": "GitHub Actions Bot",
              "fields": [
                {"title": "Status", "value": "\${{ job.status }}", "short": true},
                {"title": "Timestamp", "value": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'", "short": true}
              ]
            }' \\
            \${{ secrets.SLACK_WEBHOOK_URL }} || echo "Slack notification failed"`,
  },
  "database-operations": {
    name: "database-operations",
    description: "PostgreSQL database operations and migrations",
    steps: `      - name: Setup PostgreSQL client
        run: |
          sudo apt-get update
          sudo apt-get install -y postgresql-client

      - name: Run database migrations
        run: |
          if [ -d migrations ]; then
            for migration in migrations/*.sql; do
              echo "Running migration: $migration"
              psql "\${{ secrets.DATABASE_URL }}" -f "$migration"
            done
          else
            echo "No migrations directory found"
          fi

      - name: Create database backup
        run: |
          BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
          pg_dump "\${{ secrets.DATABASE_URL }}" > "$BACKUP_FILE"
          echo "Database backup created: $BACKUP_FILE"

      - name: Execute database query
        run: |
          psql "\${{ secrets.DATABASE_URL }}" -c "SELECT version();"
          psql "\${{ secrets.DATABASE_URL }}" -c "SELECT COUNT(*) as table_count FROM information_schema.tables WHERE table_schema = 'public';"`,
  },
  custom: {
    name: "custom-command",
    description: "Create a custom command from scratch",
    steps: `      - name: Custom step
        run: |
          echo "Add your custom commands here"
          
          # Example: Install dependencies
          # sudo apt-get update && sudo apt-get install -y your-package
          
          # Example: Run custom script
          # ./scripts/custom-script.sh
          
          # Example: Set environment variables
          # echo "CUSTOM_VAR=value" >> $GITHUB_ENV`,
  },
};

const formSchema = z.object({
  template: z.string().min(1, "Please select a template"),
  name: z
    .string()
    .min(1, "Command name is required")
    .max(50, "Name must be 50 characters or less"),
  description: z
    .string()
    .min(1, "Description is required")
    .max(200, "Description must be 200 characters or less"),
});

type FormData = z.infer<typeof formSchema>;

interface AddCommandModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAddCommand: (commandYaml: string) => void;
}

const AddCommandModal: React.FC<AddCommandModalProps> = ({
  open,
  onOpenChange,
  onAddCommand,
}) => {
  const [selectedTemplate, setSelectedTemplate] = useState<string>("");

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      template: "",
      name: "",
      description: "",
    },
  });

  const handleTemplateChange = (templateKey: string) => {
    setSelectedTemplate(templateKey);
    const template =
      COMMAND_TEMPLATES[templateKey as keyof typeof COMMAND_TEMPLATES];
    if (template) {
      form.setValue("template", templateKey);
      form.setValue("name", template.name);
      form.setValue("description", template.description);
    }
  };

  const onSubmit = (data: FormData) => {
    const template =
      COMMAND_TEMPLATES[data.template as keyof typeof COMMAND_TEMPLATES];
    if (template) {
      // Since we now provide only steps, pass them directly
      const commandSteps = template.steps;

      onAddCommand(commandSteps);

      // Reset form and close modal
      form.reset();
      setSelectedTemplate("");
      onOpenChange(false);
    }
  };

  const handleClose = () => {
    form.reset();
    setSelectedTemplate("");
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add Command</DialogTitle>
          <DialogDescription>
            Choose a template and customize the command for your skill
            configuration.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <FormField
              control={form.control}
              name="template"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Command Template</FormLabel>
                  <Select
                    onValueChange={handleTemplateChange}
                    value={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a command template" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="deploy-application">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">
                            Deploy Application
                          </span>
                          <span className="text-xs text-muted-foreground">
                            Complete deployment with GCP, database, and
                            notifications
                          </span>
                        </div>
                      </SelectItem>
                      <SelectItem value="gcloud-operations">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">GCloud Operations</span>
                          <span className="text-xs text-muted-foreground">
                            Google Cloud Platform resource management
                          </span>
                        </div>
                      </SelectItem>
                      <SelectItem value="kubectl-deployment">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">
                            Kubernetes Deployment
                          </span>
                          <span className="text-xs text-muted-foreground">
                            Deploy and manage Kubernetes clusters
                          </span>
                        </div>
                      </SelectItem>
                      <SelectItem value="python-script">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">Python Script</span>
                          <span className="text-xs text-muted-foreground">
                            Execute Python scripts with environment setup
                          </span>
                        </div>
                      </SelectItem>
                      <SelectItem value="git-operations">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">Git Operations</span>
                          <span className="text-xs text-muted-foreground">
                            Repository operations and management
                          </span>
                        </div>
                      </SelectItem>
                      <SelectItem value="ssh-deployment">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">SSH Deployment</span>
                          <span className="text-xs text-muted-foreground">
                            Deploy to remote servers via SSH
                          </span>
                        </div>
                      </SelectItem>
                      <SelectItem value="api-webhook">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">API & Webhooks</span>
                          <span className="text-xs text-muted-foreground">
                            Call external APIs and webhooks
                          </span>
                        </div>
                      </SelectItem>
                      <SelectItem value="database-operations">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">
                            Database Operations
                          </span>
                          <span className="text-xs text-muted-foreground">
                            PostgreSQL operations and migrations
                          </span>
                        </div>
                      </SelectItem>
                      <SelectItem value="custom">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">Custom Command</span>
                          <span className="text-xs text-muted-foreground">
                            Start with a basic template
                          </span>
                        </div>
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Choose a pre-built template or start with a custom command.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {selectedTemplate && (
              <>
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Command Name</FormLabel>
                      <FormControl>
                        <Input placeholder="e.g., deploy-app" {...field} />
                      </FormControl>
                      <FormDescription>
                        A unique identifier for this command (lowercase, hyphens
                        allowed).
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="description"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Description</FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder="Describe what this command does..."
                          className="min-h-[80px]"
                          {...field}
                        />
                      </FormControl>
                      <FormDescription>
                        A clear description of what this command accomplishes.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                {/* Preview of the template */}
                <div className="space-y-2">
                  <Label>Steps Preview</Label>
                  <div className="bg-muted/50 p-4 rounded-md border">
                    <pre className="text-xs overflow-x-auto whitespace-pre-wrap text-muted-foreground">
                      {
                        COMMAND_TEMPLATES[
                          selectedTemplate as keyof typeof COMMAND_TEMPLATES
                        ]?.steps
                      }
                    </pre>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    These steps will be inserted into your existing job. The
                    base job already includes checkout@v4 and runs-on:
                    ubuntu-latest.
                  </p>
                </div>
              </>
            )}

            <DialogFooter>
              <Button type="button" variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={!selectedTemplate || form.formState.isSubmitting}
              >
                Add Command
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
};

export default AddCommandModal;
