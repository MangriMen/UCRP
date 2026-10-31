import { identifierTypeLabels, type GuestProfile } from '@/entities/guest'
import { useAuth } from '@/entities/user'
import { ContactHistory } from './contact-history'

export function ProfileTab({ profile }: { profile: GuestProfile }) {
  const { can } = useAuth()

  return (
    <div className="flex flex-col gap-6">
      {can('guest.identifier.read') && (
        <section className="flex flex-col gap-3">
          <h2 className="text-lg font-semibold">Идентификаторы</h2>
          {profile.identifiers.length === 0 ? (
            <p className="text-sm text-muted-foreground">Идентификаторов нет</p>
          ) : (
            <ul className="flex flex-col gap-1 text-sm">
              {profile.identifiers.map((identifier) => (
                <li key={identifier.id}>
                  <span className="text-muted-foreground">
                    {identifierTypeLabels[identifier.idType]}:{' '}
                  </span>
                  {identifier.valueNormalized}
                </li>
              ))}
            </ul>
          )}
        </section>
      )}
      {can('guest.contact_history.read') && <ContactHistory guestId={profile.id} />}
    </div>
  )
}
