import { Toaster } from 'sonner';

import { ApiProvider } from '@/lib/api';
import { BaseXOps } from '@/pages/BaseXOps';

export function App() {
  return (
    <ApiProvider>
      <BaseXOps />
      <Toaster position="bottom-right" richColors theme="dark" />
    </ApiProvider>
  );
}
