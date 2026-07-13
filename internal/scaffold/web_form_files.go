package scaffold

import (
	"path/filepath"
)

// writeWebFormFiles scaffolds the public form sharing UI into the Next.js web app.
// It outputs to apps/web/app/forms/[token]/page.tsx.
func writeWebFormFiles(root string, opts Options) error {
	webRoot := filepath.Join(root, "apps", "web")
	
	files := map[string]string{
		filepath.Join(webRoot, "app", "forms", "[token]", "page.tsx"): stripUseClient(webFormPage()),
	}

	for path, content := range files {
		if err := writeFile(path, content); err != nil {
			return err
		}
	}
	return nil
}

func webFormPage() string {
	return `"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useQuery, useMutation } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";

interface FormShare {
  resource_name: string;
  has_password: boolean;
  label: string;
  custom_title?: string;
  custom_description?: string;
  fields: Array<{
    key: string;
    label: string;
    type: string;
    required: boolean;
  }>;
}

export default function FormSharePage() {
  const { token } = useParams() as { token: string };
  const router = useRouter();

  const [password, setPassword] = useState("");
  const [formData, setFormData] = useState<Record<string, any>>({});
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [submitError, setSubmitError] = useState("");

  const { data, isLoading, error } = useQuery({
    queryKey: ["formShare", token],
    queryFn: async () => {
      const res = await apiClient.get('/api/public/forms/' + token);
      return res.data.data as FormShare;
    },
    retry: false,
  });

  const submitMutation = useMutation({
    mutationFn: async (dataToSubmit: Record<string, any>) => {
      const res = await apiClient.post('/api/public/forms/' + token + '/submit', {
        _password: password,
        fields: dataToSubmit,
      });
      return res.data;
    },
    onSuccess: () => {
      setIsSubmitted(true);
      setSubmitError("");
    },
    onError: (err: any) => {
      setSubmitError(err.response?.data?.error?.message || "Submission failed. Please check your inputs.");
    },
  });

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-bg-primary">
        <div className="text-text-secondary">Loading form...</div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-bg-primary">
        <div className="rounded-xl border border-border bg-bg-secondary p-8 text-center max-w-md">
          <h1 className="mb-2 text-xl font-medium text-foreground">Form Not Found</h1>
          <p className="text-text-secondary">
            This link is invalid, has expired, or the form is disabled.
          </p>
        </div>
      </div>
    );
  }

  if (isSubmitted) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-bg-primary p-4">
        <div className="w-full max-w-md rounded-xl border border-border bg-bg-secondary p-8 text-center">
          <div className="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-full bg-success/20 text-success">
            <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="mb-2 text-2xl font-semibold text-foreground">Thank You</h2>
          <p className="text-text-secondary">Your submission has been received.</p>
          <Button className="mt-8 w-full" onClick={() => {
            setIsSubmitted(false);
            setFormData({});
          }}>
            Submit another response
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen justify-center bg-bg-primary px-4 py-12 sm:px-6 lg:px-8">
      <div className="w-full max-w-xl">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold tracking-tight text-foreground">
            {data.custom_title || "Submit " + data.label}
          </h1>
          {data.custom_description && (
            <p className="mt-2 text-text-secondary">{data.custom_description}</p>
          )}
        </div>

        <form 
          className="space-y-6 rounded-xl border border-border bg-bg-secondary p-6 sm:p-8"
          onSubmit={(e) => {
            e.preventDefault();
            submitMutation.mutate(formData);
          }}
        >
          {data.has_password && (
            <div className="mb-6 rounded-lg bg-bg-tertiary p-4">
              <label className="mb-1 block text-sm font-medium text-foreground">
                Password required
              </label>
              <Input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter form password"
              />
            </div>
          )}

          {data.fields.map((field) => (
            <div key={field.key} className="space-y-1">
              <label className="block text-sm font-medium text-foreground">
                {field.label} {field.required && <span className="text-danger">*</span>}
              </label>
              {field.type === "textarea" ? (
                <textarea
                  required={field.required}
                  value={formData[field.key] || ""}
                  onChange={(e) => setFormData({ ...formData, [field.key]: e.target.value })}
                  className="w-full min-h-[100px] rounded-lg border border-border bg-bg-tertiary px-3 py-2 text-foreground focus:border-accent focus:outline-none focus:ring-1 focus:ring-accent"
                />
              ) : field.type === "checkbox" ? (
                <input
                  type="checkbox"
                  required={field.required}
                  checked={!!formData[field.key]}
                  onChange={(e) => setFormData({ ...formData, [field.key]: e.target.checked })}
                  className="h-4 w-4 rounded border-border bg-bg-tertiary text-accent focus:ring-accent"
                />
              ) : (
                <Input
                  type={field.type === "number" ? "number" : field.type === "email" ? "email" : field.type === "tel" ? "tel" : "text"}
                  required={field.required}
                  value={formData[field.key] || ""}
                  onChange={(e) => setFormData({ ...formData, [field.key]: e.target.value })}
                  className="w-full"
                />
              )}
            </div>
          ))}

          {submitError && (
            <div className="rounded-lg bg-danger/10 p-3 text-sm text-danger">
              {submitError}
            </div>
          )}

          <Button 
            type="submit" 
            className="w-full" 
            disabled={submitMutation.isPending}
          >
            {submitMutation.isPending ? "Submitting..." : "Submit"}
          </Button>
        </form>
      </div>
    </div>
  );
}
`
}
