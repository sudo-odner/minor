import React, { useState } from 'react';

interface MessageInputProps {
  onSend: (content: string) => void;
  placeholder: string;
}

const MessageInput: React.FC<MessageInputProps> = ({ onSend, placeholder }) => {
  const [text, setText] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;
    onSend(text);
    setText('');
  };

  return (
    <form onSubmit={handleSubmit} className="p-4 bg-white transition-colors duration-200">
      <div className="bg-gray-50 border border-gray-200 focus-within:border-brand-blue focus-within:ring-1 focus-within:ring-brand-blue/20 rounded-xl px-4 py-2.5 flex items-center transition-all shadow-sm">
        <button type="button" className="text-brand-blue opacity-50 hover:opacity-100 mr-3 text-2xl font-light">+</button>
        <input
          type="text"
          className="bg-transparent w-full text-gray-800 focus:outline-none text-[15px] placeholder-gray-400"
          placeholder={placeholder}
          value={text}
          onChange={(e) => setText(e.target.value)}
        />
      </div>
    </form>
  );
};

export default MessageInput;