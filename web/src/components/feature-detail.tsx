import { Lock, WifiOff, Cpu } from "lucide-react";

export function FeatureDetail() {
  return (
    <section className="py-32 px-4 sm:px-6 lg:px-8 border-b border-border bg-gray-50 dark:bg-black/50">
      <div className="max-w-7xl mx-auto">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-24 items-center">
          <div>
            <h2 className="text-4xl sm:text-5xl font-bold mb-8 font-mono uppercase tracking-tighter leading-tight">
              Your Data.<br />
              Your Device.<br />
              <span className="text-gray-400">Your Rules.</span>
            </h2>
            
            <div className="space-y-12">
              <div className="flex gap-6">
                <div className="shrink-0 w-12 h-12 border border-black dark:border-white flex items-center justify-center bg-white dark:bg-black">
                  <Lock className="w-6 h-6" />
                </div>
                <div>
                  <h3 className="text-xl font-bold mb-2 font-mono uppercase">Privacy First</h3>
                  <p className="text-gray-600 dark:text-gray-400 leading-relaxed">
                    We don't train on your data. Everything is stored locally in SQLite. 
                    Cloud sync is completely optional and end-to-end encrypted.
                  </p>
                </div>
              </div>

              <div className="flex gap-6">
                <div className="shrink-0 w-12 h-12 border border-black dark:border-white flex items-center justify-center bg-white dark:bg-black">
                  <WifiOff className="w-6 h-6" />
                </div>
                <div>
                  <h3 className="text-xl font-bold mb-2 font-mono uppercase">Works Offline</h3>
                  <p className="text-gray-600 dark:text-gray-400 leading-relaxed">
                    No internet? No problem. Seisami is a native app that works perfectly without a connection.
                    Sync happens automatically when you're back online.
                  </p>
                </div>
              </div>

              <div className="flex gap-6">
                <div className="shrink-0 w-12 h-12 border border-black dark:border-white flex items-center justify-center bg-white dark:bg-black">
                  <Cpu className="w-6 h-6" />
                </div>
                <div>
                  <h3 className="text-xl font-bold mb-2 font-mono uppercase">Local AI</h3>
                  <p className="text-gray-600 dark:text-gray-400 leading-relaxed">
                    Powered by local models for transcription and intent recognition. 
                    Fast, private, and free to use.
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div className="relative aspect-square lg:aspect-auto lg:h-full min-h-[500px] border border-black dark:border-white bg-white dark:bg-black p-8 flex items-center justify-center overflow-hidden">
            <div className="absolute inset-0 opacity-10 bg-[radial-gradient(#000_1px,transparent_1px)] [background-size:16px_16px] dark:invert" />
            
            <div className="relative z-10 text-center">
              <div className="inline-block border-2 border-black dark:border-white p-8 mb-8 bg-white dark:bg-black">
                <code className="text-sm font-mono">
                  <span className="text-gray-500">// Local Storage</span><br/>
                  ~/.seisami/data.db<br/><br/>
                  <span className="text-gray-500">// Encryption</span><br/>
                  AES-256-GCM<br/><br/>
                  <span className="text-gray-500">// Sync Status</span><br/>
                  <span className="text-green-500">●</span> Offline Ready
                </code>
              </div>
              <p className="font-mono text-sm uppercase tracking-widest">System Status: Secure</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
