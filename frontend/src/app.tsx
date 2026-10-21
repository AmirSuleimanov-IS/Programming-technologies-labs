import { BrowserRouter, Navigate, Route, Routes } from 'react-router'

import { AppLayout } from './components/elements/app-layout'
import { CompetitionsPage } from './pages/competitions-page'
import { NotFoundPage } from './pages/not-found-page'

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<Navigate to="/competitions" replace />} />
          <Route path="competitions" element={<CompetitionsPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
