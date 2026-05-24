package authz

type Permission uint64

const (
	// Текстовые права

	// PermReadChat Позволяет просматривать чат
	PermReadChat Permission = 1 << 0

	// PermSendMessages Позволяет отправлять любой контент: текст, вложения, глосовые
	PermSendMessages Permission = 1 << 2

	// PermModerateChat Позволяет управлять чатам: удалять сообщения(в том числе не свои)
	PermModerateChat Permission = 1 << 2

	// Голосовые права

	// PermVoiceConnect Позволяет входить в голосовой канал
	PermVoiceConnect Permission = 1 << 10

	// PermVoiceSpeak Позволяет говорить в голосовом канал
	PermVoiceSpeak Permission = 1 << 11

	// PermVoiceStream Позволяет транслировать видео/экран
	PermVoiceStream Permission = 1 << 12

	// PermVoiceMute позволяет выключать голос, видео
	PermVoiceMute Permission = 1 << 13

	// Общая модерация

	// PermKickMembers позволяет выгонять или банить участников чата/сервера
	PermKickMembers Permission = 1 << 20

	// PermManageChannels позволяет создавать, удалять или переименовыввать каналы
	PermManageChannels Permission = 1 << 21
)

// Проверка прав доступа по одному из критериев
func Has(mask, perm Permission) bool {
	return (mask & perm) == perm
}

// Проверка прав доступа по всем perms(если все права есть)
func HasAll(mask Permission, perms ...Permission) bool {
	var required Permission
	for _, p := range perms {
		required |= p
	}
	return (mask & required) == required
}

// Провкрка прав доступа по всем perm(если хотябы один есть)
func HashAny(mask Permission, perms ...Permission) bool {
	for _, p := range perms {
		if (mask & p) == p {
			return true
		}
	}
	return false
}
