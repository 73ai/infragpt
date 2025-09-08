import { UseFormReturn } from "react-hook-form";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from "@/components/ui/form";
import { Checkbox } from "@/components/ui/checkbox";
import { OBSERVABILITY_STACK } from "@/lib/onboarding-constants";
import type { OnboardingFormData } from "@/lib/onboarding-constants";

interface ObservabilityStackStepProps {
  form: UseFormReturn<OnboardingFormData>;
}

export function ObservabilityStackStep({ form }: ObservabilityStackStepProps) {
  return (
    <div className="space-y-6">
      <FormField
        control={form.control}
        name="observabilityStack"
        render={() => (
          <FormItem>
            <div className="mb-4">
              <FormLabel className="text-base">
                Current Observability Stack
              </FormLabel>
              <FormDescription>
                Select the monitoring and observability tools you currently use
              </FormDescription>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {OBSERVABILITY_STACK.map((tool) => (
                <FormField
                  key={tool.value}
                  control={form.control}
                  name="observabilityStack"
                  render={({ field }) => {
                    return (
                      <FormItem
                        key={tool.value}
                        className="flex flex-row items-start space-x-3 space-y-0"
                      >
                        <FormControl>
                          <Checkbox
                            checked={field.value?.includes(tool.value)}
                            onCheckedChange={(checked) => {
                              return checked
                                ? field.onChange([...field.value, tool.value])
                                : field.onChange(
                                    field.value?.filter(
                                      (value) => value !== tool.value,
                                    ),
                                  );
                            }}
                          />
                        </FormControl>
                        <FormLabel className="text-sm font-normal cursor-pointer">
                          {tool.label}
                        </FormLabel>
                      </FormItem>
                    );
                  }}
                />
              ))}
            </div>
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );
}
