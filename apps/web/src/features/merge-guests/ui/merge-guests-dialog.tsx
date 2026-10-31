import { useState, type FormEvent } from 'react'
import { GuestStatusBadge, MIN_SEARCH_LENGTH, useGuestSearch, type Guest } from '@/entities/guest'
import { MergeErrorAlert, type MergeResult } from '@/entities/merge-candidate'
import { formatDateTime } from '@/shared/lib/format'
import { useDebouncedValue } from '@/shared/lib/hooks/use-debounced-value'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/ui/dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { Skeleton } from '@/shared/ui/skeleton'
import { useMergeGuests } from '../api/use-merge-guests'

const SEARCH_DEBOUNCE_MS = 300
const RESULTS_PAGE_SIZE = 20

interface GuestPickerProps {
  excludedGuestId: string
  selectedGuestId?: string
  onSelect: (guest: Guest) => void
}

function GuestPicker({ excludedGuestId, selectedGuestId, onSelect }: GuestPickerProps) {
  const [query, setQuery] = useState('')
  const debouncedQuery = useDebouncedValue(query.trim(), SEARCH_DEBOUNCE_MS)
  const guests = useGuestSearch({
    query: debouncedQuery,
    status: 'active',
    pageSize: RESULTS_PAGE_SIZE,
  })
  const isSearchable = debouncedQuery.length >= MIN_SEARCH_LENGTH
  const candidates = guests.data?.items.filter((item) => item.id !== excludedGuestId) ?? []

  const renderResults = () => {
    if (!isSearchable) {
      return (
        <p className="text-sm text-muted-foreground">
          Введите не меньше {MIN_SEARCH_LENGTH} символов
        </p>
      )
    }
    if (guests.isLoading) return <Skeleton className="h-16 w-full" />
    if (guests.error) return <ErrorBanner error={guests.error} />
    if (candidates.length === 0) {
      return <p className="text-sm text-muted-foreground">Гости не найдены</p>
    }
    return (
      <div className="flex max-h-72 flex-col gap-2 overflow-y-auto">
        {candidates.map((guest) => (
          <button
            key={guest.id}
            type="button"
            aria-pressed={guest.id === selectedGuestId}
            onClick={() => onSelect(guest)}
            className={cn(
              'flex flex-col gap-1 rounded-md border p-3 text-left',
              guest.id === selectedGuestId ? 'border-primary bg-muted' : 'hover:bg-muted/50',
            )}
          >
            <span className="flex items-center gap-2 text-sm font-medium">
              {guest.displayName ?? 'Без имени'}
              <GuestStatusBadge status={guest.status} />
            </span>
            <span className="text-xs text-muted-foreground">
              Арендаторов: {guest.tenantCount} · Событий: {guest.eventCount ?? 0} · Последняя
              активность: {guest.lastSeenAt ? formatDateTime(guest.lastSeenAt) : '—'}
            </span>
          </button>
        ))}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      <FormField label="Поиск гостя">
        {(field) => (
          <Input
            {...field}
            value={query}
            placeholder="Телефон, email, карта или имя"
            onChange={(event) => setQuery(event.target.value)}
          />
        )}
      </FormField>
      {renderResults()}
    </div>
  )
}

interface MergeGuestsDialogProps {
  guestId: string
  onMerged: (result: MergeResult) => void
}

export function MergeGuestsDialog({ guestId, onMerged }: MergeGuestsDialogProps) {
  const [open, setOpen] = useState(false)
  const [selectedGuest, setSelectedGuest] = useState<Guest | null>(null)
  const mutation = useMergeGuests()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!selectedGuest) return
    mutation.mutate(
      { guestId, targetGuestId: selectedGuest.id },
      {
        onSuccess: (result) => {
          setOpen(false)
          onMerged(result)
        },
      },
    )
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next)
        if (!next) {
          setSelectedGuest(null)
          mutation.reset()
        }
      }}
    >
      <DialogTrigger render={<Button variant="outline" />}>
        Объединить с другим профилем
      </DialogTrigger>
      <DialogContent>
        <form onSubmit={submit} className="flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle>Ручное объединение</DialogTitle>
            <DialogDescription>
              Этот профиль будет объединён в выбранный: он станет основным.
            </DialogDescription>
          </DialogHeader>
          {mutation.error && <MergeErrorAlert error={mutation.error} />}
          <GuestPicker
            excludedGuestId={guestId}
            selectedGuestId={selectedGuest?.id}
            onSelect={setSelectedGuest}
          />
          <DialogFooter>
            <Button type="submit" disabled={!selectedGuest || mutation.isPending}>
              Объединить
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
