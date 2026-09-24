#pragma once

extern void (*boleasHookInput)(void);
extern void (*boleasHookWindowReady)(void);

void boleasPrintVersion();

void boleasStartEngine(int argc, char* argv[]);

void boleasQuitEngine();
