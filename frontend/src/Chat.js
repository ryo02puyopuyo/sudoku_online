import React, { useState, useEffect, useRef } from 'react';

export default function Chat({ messages, onSendMessage }) {
  const [inputText, setInputText] = useState("");
  const messagesEndRef = useRef(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (inputText.trim()) {
      onSendMessage(inputText.trim());
      setInputText("");
    }
  };

  return (
    <div className="chat-container">
      <div className="chat-messages">
        {messages.map((msg, index) => (
          <div key={index} className="chat-message">
            <span className={`sender-name team-${msg.senderTeam}`}>{msg.senderName}</span>
            <span className="message-text">{msg.message}</span>
            <span className="message-time">{msg.timestamp}</span>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>
      <form onSubmit={handleSubmit} className="chat-input-form">
        <input
          type="text"
          value={inputText}
          onChange={(e) => setInputText(e.target.value)}
          placeholder="メッセージを入力..."
          maxLength="100"
        />
        <button type="submit">送信</button>
      </form>
    </div>
  );
}