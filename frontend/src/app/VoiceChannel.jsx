import React, { useState, useEffect, useRef, useCallback } from "react";

// Конфигурация WebRTC: используем публичный STUN-сервер Google для NAT Traversal
const pcConfig = {
  iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
};

const VoiceChannel = () => {
  // Refs для доступа к HTML Audio и глобальным объектам
  const localAudioRef = useRef(null);
  const remoteAudioRef = useRef(null);
  const pcRef = useRef(null);
  const wsRef = useRef(null);
  const localStreamRef = useRef(null);

  // Состояние UI
  const [isConnected, setIsConnected] = useState(false);
  const [status, setStatus] = useState("Инициализация...");

  // useCallback для функций, использующих глобальные Ref'ы
  const sendSignaling = useCallback((message) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  // ===================================
  // 1. Настройка PeerConnection
  // ===================================
  const setupPeerConnection = useCallback(() => {
    if (pcRef.current) return;

    const newPc = new RTCPeerConnection(pcConfig);
    pcRef.current = newPc;

    // Обработчик ICE-кандидатов (отправляем их другому пиру через WS)
    newPc.onicecandidate = (e) => {
      if (e.candidate) {
        sendSignaling({
          type: "candidate",
          candidate: e.candidate,
        });
      }
    };

    // Обработчик удаленного потока (когда приходят данные от друга)
    newPc.ontrack = (event) => {
      if (
        remoteAudioRef.current &&
        remoteAudioRef.current.srcObject !== event.streams[0]
      ) {
        remoteAudioRef.current.srcObject = event.streams[0];
        setStatus("Голосовая связь установлена!");
      }
    };

    // Отслеживание состояния соединения
    newPc.oniceconnectionstatechange = () => {
      const state = newPc.iceConnectionState;
      setStatus(`ICE State: ${state}`);
      setIsConnected(state === "connected" || state === "completed");
    };
  }, [sendSignaling]);

  // ===================================
  // 2. Обработка входящих сообщений сигнализации
  // ===================================
  const handleSignalingMessage = useCallback(
    async (message) => {
      // Если еще нет PC, создаем его (это происходит, если мы получаем Offer)
      if (!pcRef.current) {
        setupPeerConnection();
      }

      if (message.type === "offer") {
        setStatus("Получен Offer. Отправка Answer...");

        // Получаем медиа, если еще не получили
        if (!localStreamRef.current) {
          const stream = await navigator.mediaDevices.getUserMedia({
            audio: true,
          });
          localStreamRef.current = stream;
          localAudioRef.current.srcObject = stream;
          stream.getTracks().forEach((track) => {
            pcRef.current.addTrack(track, stream);
          });
        }

        // Установка Offer и создание Answer
        await pcRef.current.setRemoteDescription(
          new RTCSessionDescription(message),
        );
        const answer = await pcRef.current.createAnswer();
        await pcRef.current.setLocalDescription(answer);
        sendSignaling(answer);
      } else if (message.type === "answer") {
        // Получен Answer
        setStatus("Получен Answer. Соединение устанавливается...");
        await pcRef.current.setRemoteDescription(
          new RTCSessionDescription(message),
        );
      } else if (message.type === "candidate" && message.candidate) {
        // Получен ICE-кандидат
        try {
          await pcRef.current.addIceCandidate(
            new RTCIceCandidate(message.candidate),
          );
        } catch (e) {
          console.error("Ошибка при добавлении ICE-кандидата:", e);
        }
      }
    },
    [setupPeerConnection, sendSignaling],
  );

  // ===================================
  // 3. Установка WebSocket и логика старта
  // ===================================
  useEffect(() => {
    const connectWebSocket = () => {
      const newWs = new WebSocket("ws://localhost:8080/ws");
      wsRef.current = newWs;

      newWs.onopen = () =>
        setStatus("WebSocket подключен. Нажмите, чтобы подключиться.");

      newWs.onerror = (e) => {
        setStatus(`WebSocket Ошибка! Проверьте, запущен ли Go-сервер на 8080.`);
        console.error("WebSocket Error:", e);
      };

      newWs.onmessage = (event) => {
        try {
          handleSignalingMessage(JSON.parse(event.data));
        } catch (e) {
          console.error("Ошибка парсинга JSON:", e);
        }
      };
    };

    connectWebSocket();

    // Очистка при размонтировании
    return () => {
      if (wsRef.current) wsRef.current.close();
      if (pcRef.current) pcRef.current.close();
      if (localStreamRef.current)
        localStreamRef.current.getTracks().forEach((track) => track.stop());
    };
  }, [handleSignalingMessage]);

  // ===================================
  // 4. Запуск звонка (Offerer)
  // ===================================
  const startCallAsOfferer = async () => {
    if (
      isConnected ||
      !wsRef.current ||
      wsRef.current.readyState !== WebSocket.OPEN
    )
      return;

    try {
      // 1. Получение медиа
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      localStreamRef.current = stream;
      localAudioRef.current.srcObject = stream;

      setStatus("Микрофон получен. Отправка Offer...");

      // 2. Настройка PeerConnection и добавление потоков
      setupPeerConnection();
      stream.getTracks().forEach((track) => {
        pcRef.current.addTrack(track, stream);
      });

      // 3. Создание и отправка Offer
      const offer = await pcRef.current.createOffer();
      await pcRef.current.setLocalDescription(offer);
      sendSignaling(offer);
    } catch (e) {
      setStatus(`Ошибка доступа к микрофону: ${e.message}`);
      console.error("Ошибка доступа к медиа:", e);
    }
  };

  return (
    <div
      style={{ padding: "20px", border: "1px solid #ccc", borderRadius: "8px" }}
    >
      <h2>Голосовой Канал Go/React MVP</h2>
      <p>
        Статус: <strong>{status}</strong>
      </p>

      <button
        onClick={startCallAsOfferer}
        disabled={isConnected || wsRef.current?.readyState !== WebSocket.OPEN}
        style={{
          padding: "10px 20px",
          fontSize: "16px",
          backgroundColor: isConnected
            ? "green"
            : wsRef.current?.readyState === WebSocket.OPEN
              ? "blue"
              : "gray",
          color: "white",
          cursor: "pointer",
        }}
      >
        {isConnected ? "В канале" : "Подключиться к каналу"}
      </button>

      <hr />
      <h4>Локальное аудио (Вы):</h4>
      <audio
        ref={localAudioRef}
        autoPlay
        muted
        style={{ width: "100%" }}
      ></audio>

      <h4>Удаленное аудио (Друг):</h4>
      <audio ref={remoteAudioRef} autoPlay style={{ width: "100%" }}></audio>

      <p style={{ marginTop: "15px", fontSize: "12px" }}>
        **Инструкция:** Запустите Go-сервер (8080), откройте эту страницу в двух
        вкладках. Нажмите "Подключиться" в каждой из них.
      </p>
    </div>
  );
};

export default VoiceChannel;
