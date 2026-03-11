import { ApiProvider } from '@/lib/api';
import { BaseXOps } from '@/pages/BaseXOps';

export function App() {
  return (
    <ApiProvider>
      <BaseXOps />
    </ApiProvider>
  );
}
