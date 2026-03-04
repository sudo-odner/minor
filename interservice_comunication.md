1. Синхронное взаимодействие (gRPC)
Используется, когда сервису Б нужен ответ от сервиса А, чтобы продолжить работу. Это блокирующие вызовы, поэтому они должны быть максимально быстрыми (Low Latency).
Сценарий А: Валидация токена при подключении (Gateway -> Auth)
Когда пользователь открывает приложение, он устанавливает WebSocket-соединение с Gateway Service.
Gateway Service получает JWT токен от клиента.
Gateway делает gRPC вызов ValidateToken в Auth Service.
Auth Service проверяет подпись, срок действия и не отозван ли токен.
Auth Service возвращает user_id и базовые права.
Если ответ OK, Gateway открывает WebSocket соединение.
Сценарий Б: Проверка прав на отправку сообщения (Chat -> Guild)
Пользователь пишет сообщение в закрытый канал.
Chat Service получает HTTP/REST запрос на создание сообщения.
Chat Service должен понять: "А имеет ли юзер право писать в этот канал?". У него нет данных о ролях.
Chat Service делает gRPC вызов CheckPermissions в Guild Service.
Аргументы: user_id, guild_id, channel_id, permission_bitmask (SEND_MESSAGES).
Guild Service смотрит в Postgres (или свой Redis-кеш), вычисляет права и возвращает bool (Разрешено/Запрещено).
Если false, Chat Service сразу возвращает ошибку 403, не нагружая Kafka.
Сценарий В: Получение профиля пользователя (Guild -> User)
При отображении списка участников сервера нужно показать их ники и аватарки.
Guild Service (или API Gateway при сборке ответа) делает gRPC вызов GetUsersBatch в User Service.
Передает список [user_id_1, user_id_2, ...].
User Service достает данные из Postgres и возвращает карту Map<UserID, UserProfile>.


2. Асинхронное взаимодействие (Kafka)
Используется для уведомлений, рассылок и действий, которые могут занять время. Сервис-источник просто кидает событие ("выстрелил и забыл"), не блокируя клиента.
Сценарий Г: Отправка и доставка сообщения (Сердце чата)
Самый нагруженный сценарий. Используем паттерн "Fan-out" (рассылка многим).
Топик: chat.messages
Ключ партиционирования: channel_id (важно, чтобы порядок сообщений в канале сохранялся).
Chat Service (Producer):
Успешно сохранил сообщение в Cassandra.
Отправляет событие MessageCreated в Kafka.
Payload: { message_id, channel_id, author_id, content, mentions: [...] }.
Gateway Service (Consumer Group 1):
Читает событие.
Определяет, какие пользователи сейчас онлайн и сидят в этом канале/сервере (через Redis Session Store).
Отправляет JSON по WebSocket всем подключенным клиентам.
Notification Service (Consumer Group 2):
Читает событие.
Проверяет поле mentions (упоминания).
Если кого-то упомянули и он офлайн — формирует Push-уведомление для FCM/APNS.
Search Service (Consumer Group 3):
Читает событие.
Асинхронно индексирует текст сообщения в Elasticsearch для будущего поиска.
Сценарий Д: Изменение статуса "Онлайн" (Presence)
Пользователь запустил игру или просто зашел в сеть.
Топик: user.presence
Ключ: user_id
Gateway Service (Producer):
Обнаружил активность WebSocket или разрыв соединения.
Отправляет событие PresenceUpdate.
Payload: { user_id, status: "dnd", activity: "Playing Minecraft" }.
Presence Service (Consumer):
Обновляет данные в Redis (чтобы при запросе профиля данные были свежими).
Gateway Service (Consumer — да, он и пишет, и читает):
Это нужно для других шардов (инстансов) Gateway.
Если Вася подключен к Шарду-1, а его друг Петя к Шарду-2, Шард-2 должен узнать через Kafka, что Вася онлайн, чтобы отправить Пете зеленый кружочек.
Сценарий Е: Вступление на сервер (Guild Events)
Пользователь принял приглашение.
Топик: guild.events
Ключ: guild_id
Guild Service (Producer):
Добавил запись в Postgres.
Отправляет событие GuildMemberAdd.
Chat Service (Consumer):
Генерирует системное сообщение "Добро пожаловать, @User!" в дефолтный канал сервера.
Gateway Service (Consumer):
Рассылает всем участникам сервера пакет "User Joined", чтобы клиент обновил список участников справа.
Сценарий Ж: Обновление аватарки (User Updates)
Пользователь сменил данные профиля.
Топик: user.updates
Ключ: user_id
User Service (Producer):
Обновил запись в БД.
Отправляет событие UserUpdated.
Gateway Service (Consumer):
Это хитрая часть. Gateway должен найти все сервера и всех друзей, где этот пользователь виден, и отправить им новый URL аватарки, чтобы кэш на клиентах сбросился.

Сводная таблица потоков

Действие	                    Кто инициирует	        Кому (Метод)	            Зачем	                            Тип
Логин                           Gateway	                Auth (gRPC)	                Проверка JWT	                    Sync
Есть ли права писать?	        Chat	                Guild (gRPC)	            Валидация прав	                    Sync
Загрузка профилей	            Guild/Chat	            User (gRPC)	                Обогащение данных	                Sync
Новое сообщение	                Chat	                Kafka (chat.messages)	    Рассылка, Пуши, Поиск	            Async
Я онлайн / играю	            Gateway	                Kafka (user.presence)	    Обновление статусов	                Async
Вход/Выход с сервера	        Guild	                Kafka (guild.events)	    Обновление списков, Приветствия	    Async
Изменил аватар/ник	            User	                Kafka (user.updates)	    Инвалидация кешей	                Async
