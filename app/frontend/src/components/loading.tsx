import { Loader2 } from "lucide-react";

interface LoadingScreenProps {
  message?: string;
}

export function LoadingScreen({ message = "Loading..." }: LoadingScreenProps) {
  return (
    <div className="min-h-screen bg-neutral-50 flex items-center justify-center">
      <div className="text-center">
        <Loader2 size="lg" className="mx-auto mb-4 animate-spin" />
        <p className="text-neutral-600">{message}</p>
      </div>
    </div>
  );
}
