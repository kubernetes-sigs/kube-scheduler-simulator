import { reactive } from "@nuxtjs/composition-api";

export interface stateType {
  isOpen: boolean;
  message: string;
  messageType: MessageType;
}

export interface SnackbarPayload {
  message: string;
  messageType: MessageType;
}

export type MessageType = "info" | "error";

export default function snackbarStore() {
  const state: stateType = reactive({
    message: "",
    messageType: "error",
    isOpen: false,
  } as stateType);

  return {
    get message() {
      return state.message;
    },
    get isOpen() {
      return state.isOpen;
    },
    get messageType() {
      return state.messageType;
    },
    open() {
      state.isOpen = true;
    },
    close() {
      state.isOpen = false;
    },
    setIsOpen(isOpen: boolean) {
      if (isOpen) {
        this.open();
      } else {
        this.close();
      }
    },
    setServerErrorMessage(error: string) {
      const servererrormsg: string = "Server error occurred: ";

      state.message = servererrormsg + error;
      this.setMessageType("error");
      this.open();
    },
    setServerInfoMessage(message: string) {
      state.message = message;
      this.setMessageType("info");
      this.open();
    },
    setMessageType(messageType: MessageType) {
      state.messageType = messageType;
    },
  };
}

export type SnackBarStore = ReturnType<typeof snackbarStore>;
