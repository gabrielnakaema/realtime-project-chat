import { createContext, useContext, useState } from 'react';

type MessagesSheetView = 'list' | 'compose' | 'chat';

interface MessagesSheetContextData {
  isOpen: boolean;
  selectedChatId: string | null;
  view: MessagesSheetView;
  open: () => void;
  openCompose: () => void;
  openChat: (chatId: string) => void;
  close: () => void;
  backToList: () => void;
}

export const MessagesSheetContext = createContext<MessagesSheetContextData>({} as MessagesSheetContextData);

export const MessagesSheetProvider = ({ children }: { children: React.ReactNode }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null);
  const [view, setView] = useState<MessagesSheetView>('list');

  const open = () => {
    setSelectedChatId(null);
    setView('list');
    setIsOpen(true);
  };

  const openCompose = () => {
    setSelectedChatId(null);
    setView('compose');
    setIsOpen(true);
  };

  const openChat = (chatId: string) => {
    setSelectedChatId(chatId);
    setView('chat');
    setIsOpen(true);
  };

  const close = () => {
    setIsOpen(false);
    setSelectedChatId(null);
    setView('list');
  };

  const backToList = () => {
    setSelectedChatId(null);
    setView('list');
  };

  return (
    <MessagesSheetContext.Provider value={{ isOpen, selectedChatId, view, open, openCompose, openChat, close, backToList }}>
      {children}
    </MessagesSheetContext.Provider>
  );
};

export const useMessagesSheet = () => useContext(MessagesSheetContext);
