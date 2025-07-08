export const COMPANY_SIZES = [
  { value: 'startup', label: 'Startup (1-10 employees)' },
  { value: 'small', label: 'Small (11-50 employees)' },
  { value: 'medium', label: 'Medium (51-200 employees)' },
  { value: 'large', label: 'Large (201-1000 employees)' },
  { value: 'enterprise', label: 'Enterprise (1000+ employees)' },
] as const;

export const TEAM_SIZES = [
  { value: '1-5', label: '1-5 people' },
  { value: '6-20', label: '6-20 people' },
  { value: '21-50', label: '21-50 people' },
  { value: '51-100', label: '51-100 people' },
  { value: '100+', label: '100+ people' },
] as const;

export const USE_CASES = [
  { value: 'infrastructure_monitoring', label: 'Infrastructure Monitoring' },
  { value: 'application_performance_monitoring', label: 'Application Performance Monitoring' },
  { value: 'log_management', label: 'Log Management' },
  { value: 'incident_response', label: 'Incident Response' },
  { value: 'compliance_auditing', label: 'Compliance & Auditing' },
  { value: 'cost_optimization', label: 'Cost Optimization' },
  { value: 'security_monitoring', label: 'Security Monitoring' },
  { value: 'devops_automation', label: 'DevOps Automation' },
] as const;

export const OBSERVABILITY_STACK = [
  { value: 'datadog', label: 'Datadog' },
  { value: 'new_relic', label: 'New Relic' },
  { value: 'splunk', label: 'Splunk' },
  { value: 'elastic_stack', label: 'Elastic Stack' },
  { value: 'prometheus_grafana', label: 'Prometheus + Grafana' },
  { value: 'app_dynamics', label: 'AppDynamics' },
  { value: 'dynatrace', label: 'Dynatrace' },
  { value: 'cloudwatch', label: 'CloudWatch' },
  { value: 'azure_monitor', label: 'Azure Monitor' },
  { value: 'google_cloud_monitoring', label: 'Google Cloud Monitoring' },
  { value: 'pagerduty', label: 'PagerDuty' },
  { value: 'opsgenie', label: 'Opsgenie' },
  { value: 'other', label: 'Other' },
] as const;

export type CompanySize = typeof COMPANY_SIZES[number]['value'];
export type TeamSize = typeof TEAM_SIZES[number]['value'];
export type UseCase = typeof USE_CASES[number]['value'];
export type ObservabilityStack = typeof OBSERVABILITY_STACK[number]['value'];

export interface OnboardingFormData {
  companySize: CompanySize;
  teamSize: TeamSize;
  useCases: UseCase[];
  observabilityStack: ObservabilityStack[];
}