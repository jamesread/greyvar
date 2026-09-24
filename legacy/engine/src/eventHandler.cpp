#include "boleas.hpp"
#include "util.hpp"                                                           
#include "Renderer.hpp"                                                         
#include "GameState.hpp"                                                        
#include "gui/Gui.hpp"                                                          
#include "Startup.hpp"                                                          
#include "YamlNode.hpp"                                                         
#include "cvars.hpp"                                                            
#include "input/0_hid/common.hpp"                                               
#include "input/1_bindings/common.hpp"                                          
#include "input/2_actions/common.hpp"                                           
#include "LTimer.hpp" 

#define MAX_FPS 30
#define TICKS_PER_FRAME (1000 / MAX_FPS)

bool doTheLoopyLoop = true;

void eventHandler() {                                                           
    SDL_Event e{};                                                              
                                                                                
    while (SDL_PollEvent(&e) != 0) {                                            
        switch (e.type) {                                                       
            case SDL_TEXTINPUT:                                                 
            case SDL_TEXTEDITING:                                               
            case SDL_MOUSEMOTION:                                               
                // Mouse movement is handled outside of the input system, because
                // there can be a billion events just by dragging the cursor    
                // and it's only relevant for the menuscreens                   
                Gui::get().onMouseMoved(e.motion.x, e.motion.y);                
                break;                                                          
            case SDL_MOUSEWHEEL:                                                
            case SDL_MOUSEBUTTONDOWN:                                           
            case SDL_JOYBUTTONDOWN: // we use the controller api                
            case SDL_JOYBUTTONUP:                                               
            case SDL_JOYDEVICEADDED:                                            
            case SDL_JOYAXISMOTION: // ^^                                       
                break;                                                          
                                                                                
            case SDL_MOUSEBUTTONUP:                                             
                recvMouseButtonClicked(e);                                      
                break;                                                          
            case SDL_CONTROLLERBUTTONDOWN:                                      
                recvGamepadButtonDown(e);                                       
                break;                                                          
                                                                                
            case SDL_WINDOWEVENT:                                               
                switch (e.window.event) {                                       
                    case SDL_WINDOWEVENT_SHOWN:                                 
                    case SDL_WINDOWEVENT_RESIZED:                               
                    case SDL_WINDOWEVENT_SIZE_CHANGED:                          
                        SDL_GetWindowSize(Renderer::get().getWindow(), &Renderer::get().window_w, &Renderer::get().window_h);
                                                                                
                        Gui::get().onWindowResized();                           
                                                                                
                        break;                                                  
                    case SDL_WINDOWEVENT_MINIMIZED:                             
                        cout << "min" << e.window.type << endl;                 
                    case SDL_WINDOWEVENT_ENTER:                                 
                    case SDL_WINDOWEVENT_LEAVE:                                 
                        break;                                                  
                    default:                                                    
                        break;                                                  
                }                                                               
                                                                                
                break;                                                          
            case SDL_QUIT:                                                      
                doTheLoopyLoop = false;                                         
                return;                                                         
            case SDL_CONTROLLERAXISMOTION:                                      
                // poll instead                                                 
                break;                                                          
            case SDL_KEYDOWN:                                                   
                recvKeydownInput(e);                                            
                break;                                                          
            case SDL_KEYUP:                                                     
                recvKeyupInput(e);                                              
                break;                                                          
            case SDL_AUDIODEVICEADDED:                                          
                std::cout << "Audio device found: " << SDL_GetAudioDeviceName(e.adevice.which, e.adevice.iscapture) << std::endl;
                break;                                                          
            case SDL_CONTROLLERDEVICEREMOVED:                                   
                removeGamepad(e.cdevice.which);                                 
                break;                                                          
            case SDL_CONTROLLERDEVICEADDED:                                     
                initGamepad(e.cdevice.which);                                   
                break;                                                          
            default:                                                            
                std::cout << "Unknown event: 0x" << hex << e.type << std::endl; 
        }                                                                       
    }                                                                           
}   


void mainLoopRecvInput() {
    resetKeyboard();

    eventHandler();

    recvKeyboardInput();
    recvGamepadInputs();
}

void mainLoopProcess() {
	boleasHookInput();

    lookupActionBindingForPlayerInput();
    executeActionInputs();
}

void mainLoopRender(int countedFrames, int ticks) {
    Renderer::get().renderFrame();
    Renderer::get().averageFps = countedFrames / (ticks / 1000.f);
}

void mainLoop() {
    int countedFrames = 0;

    LTimer fps;
    fps.start();

    LTimer cap;

    while (doTheLoopyLoop) {
        cap.start();

        mainLoopRecvInput();
        mainLoopProcess();
        mainLoopRender(countedFrames, fps.getTicks());

        countedFrames++;

        int frameTicks = cap.getTicks();

        if (frameTicks < TICKS_PER_FRAME) {
            SDL_Delay(TICKS_PER_FRAME - frameTicks);
        }
    }
}

void pushSdlQuit() {
    std::cout << "pushSdlQuit()" << endl;

    SDL_Event e{};
    e.type = SDL_QUIT;

    SDL_PushEvent(&e);
}
