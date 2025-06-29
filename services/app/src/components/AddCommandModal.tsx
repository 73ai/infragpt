import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';

// Command templates for common GitHub Actions patterns
const COMMAND_TEMPLATES = {
  'ci-build': {
    name: 'ci-build',
    description: 'Build and test the project using CI/CD pipeline',
    template: `  - name: "ci-build"
    description: "Build and test the project using CI/CD pipeline"
    parameters:
      - name: "branch"
        type: "string"
        required: false
        description: "Branch to build (defaults to main)"
        default: "main"
      - name: "environment"
        type: "string"
        required: false
        description: "Target environment"
        default: "staging"
    actions:
      - uses: actions/checkout@v4
        with:
          ref: \${{ parameters.branch }}
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
      - run: npm ci
      - run: npm run build
      - run: npm test`
  },
  'deploy-service': {
    name: 'deploy-service',
    description: 'Deploy service to cloud infrastructure',
    template: `  - name: "deploy-service"
    description: "Deploy service to cloud infrastructure"
    parameters:
      - name: "service-name"
        type: "string"
        required: true
        description: "Name of the service to deploy"
      - name: "environment"
        type: "string"
        required: true
        description: "Target environment (dev, staging, prod)"
      - name: "image-tag"
        type: "string"
        required: false
        description: "Docker image tag"
        default: "latest"
    actions:
      - uses: actions/checkout@v4
      - name: Configure credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: \${{ secrets.AWS_ROLE_ARN }}
      - name: Deploy to ECS
        run: |
          aws ecs update-service \\
            --cluster \${{ parameters.environment }} \\
            --service \${{ parameters.service-name }} \\
            --force-new-deployment`
  },
  'run-tests': {
    name: 'run-tests',
    description: 'Execute test suite with coverage reporting',
    template: `  - name: "run-tests"
    description: "Execute test suite with coverage reporting"
    parameters:
      - name: "test-type"
        type: "string"
        required: false
        description: "Type of tests to run (unit, integration, e2e)"
        default: "unit"
      - name: "coverage-threshold"
        type: "number"
        required: false
        description: "Minimum coverage percentage required"
        default: 80
    actions:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
      - run: npm ci
      - name: Run tests
        run: npm run test:\${{ parameters.test-type }} -- --coverage
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          token: \${{ secrets.CODECOV_TOKEN }}`
  },
  'database-migration': {
    name: 'database-migration',
    description: 'Run database migrations safely',
    template: `  - name: "database-migration"
    description: "Run database migrations safely"
    parameters:
      - name: "migration-direction"
        type: "string"
        required: false
        description: "Migration direction (up, down)"
        default: "up"
      - name: "environment"
        type: "string"
        required: true
        description: "Target environment"
      - name: "dry-run"
        type: "boolean"
        required: false
        description: "Run migration in dry-run mode"
        default: false
    actions:
      - uses: actions/checkout@v4
      - name: Setup database connection
        run: |
          echo "DB_URL=\${{ secrets.DATABASE_URL_\${{ upper(parameters.environment) }} }}" >> $GITHUB_ENV
      - name: Run migrations
        run: |
          if [ "\${{ parameters.dry-run }}" = "true" ]; then
            npm run migrate:dry-run
          else
            npm run migrate:\${{ parameters.migration-direction }}
          fi`
  },
  'security-scan': {
    name: 'security-scan',
    description: 'Perform security vulnerability scanning',
    template: `  - name: "security-scan"
    description: "Perform security vulnerability scanning"
    parameters:
      - name: "scan-type"
        type: "string"
        required: false
        description: "Type of security scan (dependencies, code, container)"
        default: "dependencies"
      - name: "fail-on-high"
        type: "boolean"
        required: false
        description: "Fail build on high severity vulnerabilities"
        default: true
    actions:
      - uses: actions/checkout@v4
      - name: Dependency scan
        if: parameters.scan-type == 'dependencies'
        run: npm audit --audit-level=\${{ parameters.fail-on-high && 'high' || 'critical' }}
      - name: Code scan
        if: parameters.scan-type == 'code'
        uses: github/codeql-action/analyze@v3
      - name: Container scan
        if: parameters.scan-type == 'container'
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: \${{ env.IMAGE_NAME }}
          format: 'sarif'
          output: 'trivy-results.sarif'`
  },
  'custom': {
    name: 'custom-command',
    description: 'Create a custom command from scratch',
    template: `  - name: "custom-command"
    description: "Custom command description"
    parameters:
      - name: "example-param"
        type: "string"
        required: true
        description: "Example parameter description"
    actions:
      - uses: actions/checkout@v4
      - name: Custom step
        run: echo "Add your custom commands here"`
  }
};

const formSchema = z.object({
  template: z.string().min(1, 'Please select a template'),
  name: z.string().min(1, 'Command name is required').max(50, 'Name must be 50 characters or less'),
  description: z.string().min(1, 'Description is required').max(200, 'Description must be 200 characters or less'),
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
  const [selectedTemplate, setSelectedTemplate] = useState<string>('');

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      template: '',
      name: '',
      description: '',
    },
  });

  const handleTemplateChange = (templateKey: string) => {
    setSelectedTemplate(templateKey);
    const template = COMMAND_TEMPLATES[templateKey as keyof typeof COMMAND_TEMPLATES];
    if (template) {
      form.setValue('template', templateKey);
      form.setValue('name', template.name);
      form.setValue('description', template.description);
    }
  };

  const onSubmit = (data: FormData) => {
    const template = COMMAND_TEMPLATES[data.template as keyof typeof COMMAND_TEMPLATES];
    if (template) {
      let commandYaml = template.template;
      
      // Replace name and description if they were customized
      if (data.name !== template.name) {
        commandYaml = commandYaml.replace(
          `name: "${template.name}"`,
          `name: "${data.name}"`
        );
      }
      
      if (data.description !== template.description) {
        commandYaml = commandYaml.replace(
          `description: "${template.description}"`,
          `description: "${data.description}"`
        );
      }

      onAddCommand(commandYaml);
      
      // Reset form and close modal
      form.reset();
      setSelectedTemplate('');
      onOpenChange(false);
    }
  };

  const handleClose = () => {
    form.reset();
    setSelectedTemplate('');
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add Command</DialogTitle>
          <DialogDescription>
            Choose a template and customize the command for your skill configuration.
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
                      <SelectItem value="ci-build">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">CI Build</span>
                          <span className="text-xs text-muted-foreground">Build and test project</span>
                        </div>
                      </SelectItem>
                      <SelectItem value="deploy-service">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">Deploy Service</span>
                          <span className="text-xs text-muted-foreground">Deploy to cloud infrastructure</span>
                        </div>
                      </SelectItem>
                      <SelectItem value="run-tests">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">Run Tests</span>
                          <span className="text-xs text-muted-foreground">Execute test suite with coverage</span>
                        </div>
                      </SelectItem>
                      <SelectItem value="database-migration">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">Database Migration</span>
                          <span className="text-xs text-muted-foreground">Run database migrations safely</span>
                        </div>
                      </SelectItem>
                      <SelectItem value="security-scan">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">Security Scan</span>
                          <span className="text-xs text-muted-foreground">Perform vulnerability scanning</span>
                        </div>
                      </SelectItem>
                      <SelectItem value="custom">
                        <div className="flex flex-col items-start">
                          <span className="font-medium">Custom Command</span>
                          <span className="text-xs text-muted-foreground">Start with a basic template</span>
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
                        <Input 
                          placeholder="e.g., deploy-app" 
                          {...field} 
                        />
                      </FormControl>
                      <FormDescription>
                        A unique identifier for this command (lowercase, hyphens allowed).
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
                  <Label>Template Preview</Label>
                  <div className="bg-muted/50 p-4 rounded-md border">
                    <pre className="text-xs overflow-x-auto whitespace-pre-wrap text-muted-foreground">
                      {COMMAND_TEMPLATES[selectedTemplate as keyof typeof COMMAND_TEMPLATES]?.template}
                    </pre>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    This template will be inserted into your YAML editor. You can customize it further after adding.
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