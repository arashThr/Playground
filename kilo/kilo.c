#include <unistd.h>
#include <stdio.h>
#include <ctype.h>
#include <termios.h>
#include <stdlib.h>
#include <errno.h>
#include <sys/ioctl.h>
#include <string.h>
#include <sys/types.h>

#define KILO_VERSION "0.0.1"

#define CTRL_KEY(k) ((k) & 0x1F)

enum editorKey {
    ARROW_LEFT = 1000,
    ARROW_RIGHT,
    ARROW_UP,
    ARROW_DOWN,
    HOME_KEY,
    END_KEY,
    PAGE_DOWN,
    PAGE_UP,
    DEL_KEY,
};

typedef struct {
    char* line;
    int len;
} erow;

struct editorConfigs {
    int cx, cy;
    int screenRows;
    int screenCols;
    int numRows;
    erow** rows;
    int rowOff;
    int colOff;
    struct termios orig_term;
};

struct editorConfigs E;

/** Terminal **/

void die(const char* err) {
    write(STDOUT_FILENO, "\x1b[2J", 4);
    write(STDOUT_FILENO, "\x1b[H", 3);

    perror(err);
    exit(1);
}

void disableRawMode() {
    if (tcsetattr(STDIN_FILENO, TCSAFLUSH, &E.orig_term) == -1)
        die("tcsetattr");
}

void enableRawMode() {
    if (tcgetattr(STDIN_FILENO, &E.orig_term) == -1)
        die("tcgetattr");

    struct termios raw = E.orig_term;
    atexit(disableRawMode);

    // IXON: Ctrl-S Ctrl-Q - Disable "Software control flow": Stop sending chars to the teminal (for old days)
    // ICRNL: Ctrl-M - Translate \r to new line
    raw.c_iflag &= ~(IXON | ICRNL);

    // Output processing: \n to \r\n
    // From now on, new line need both
    raw.c_oflag &= ~(OPOST);

    // IEXTEN: Ctrl-V
    // ISIG: Ctrl-Z Ctrl-C
    raw.c_lflag &= ~(ECHO | ICANON | ISIG | IEXTEN);

    // Misc
    raw.c_cflag |= (CS8);
    raw.c_iflag &= ~(BRKINT | INPCK | ISTRIP);

    // Min number of bytes before read returns
    raw.c_cc[VMIN] = 0;
    // Max time before read returns: tenth of second
    raw.c_cc[VTIME] = 10;

    if (tcsetattr(STDIN_FILENO, TCSAFLUSH, &raw) == -1)
        die("tcsetattr");
}

int editorReadKey() {
    char c = '\0';
    int nread = 0;
    while ((nread = read(STDIN_FILENO, &c, 1)) != 1) {
        if (nread == -1 && errno != EAGAIN)
            die("read");
    }

    if (c == '\x1b') {
        char seq[3];

        if (read(STDOUT_FILENO, &seq[0], 1) != 1) return '\x1b';
        if (read(STDOUT_FILENO, &seq[1], 1) != 1) return '\x1b';

        if (seq[0] == '[') {
            if (seq[1] >= '0' && seq[1] <= '9') {
                if (read(STDIN_FILENO, &seq[2], 1) != 1) return '\x1b';
                if (seq[2] == '~') {
                    switch (seq[1]) {
                        case '1': return HOME_KEY;
                        case '3': return DEL_KEY;
                        case '4': return END_KEY;
                        case '5': return PAGE_UP;
                        case '6': return PAGE_DOWN;
                        case '7': return HOME_KEY;
                        case '8': return END_KEY;
                    }
                }
            } else {
                switch (seq[1]) {
                    case 'A': return ARROW_UP;
                    case 'B': return ARROW_DOWN;
                    case 'C': return ARROW_RIGHT;
                    case 'D': return ARROW_LEFT;
                    case 'H': return HOME_KEY;
                    case 'F': return END_KEY;
                }
            }
        } else if (seq[0] == 'O') {
            switch (seq[1]) {
                case 'H': return HOME_KEY;
                case 'F': return END_KEY;
            }
        }
        return '\x1b';
    }
    return c;
}

int getWindowSize(int* rows, int* cols) {
    struct winsize ws;

    if (ioctl(STDOUT_FILENO, TIOCGWINSZ, &ws) == -1 || ws.ws_col == 0) {
        return -1;
    } else {
        *cols = ws.ws_col;
        *rows = ws.ws_row;
        return 0;
    }
}

/** Buffer **/

struct abuf {
    char* b;
    int len;
};

#define ABUF_INIT { NULL, 0 }

void abAppend(struct abuf* ab, char* s, int len) {
    char* new = reallocf(ab->b, ab->len + len);

    if (new == NULL)
        return;
    memcpy(&new[ab->len], s, len);
    ab->b = new;
    ab->len += len;
}

void abFree(struct abuf* ab) {
    free(ab->b);
}

/** Input **/

void editMoveCursor(int key) {
    switch (key) {
        case ARROW_UP:
            {
                if (E.cy > 0) E.cy -= 1;
                else if (E.rowOff > 0) E.rowOff -= 1; // Scroll up

                int nextLineLen = E.rows[E.rowOff + E.cy]->len;
                if (nextLineLen < E.cx) E.cx = nextLineLen;
                else E.cx = nextLineLen < E.colOff ? nextLineLen : E.colOff;
            }
            break;
        case ARROW_LEFT:
            if (E.cx > 0) {
                E.cx -= 1;
                E.colOff = E.cx;
            }
            break;
        case ARROW_DOWN:
            {
                if (E.rowOff + E.cy >= E.numRows) break; // End of the file
                if (E.cy < E.screenRows - 1) E.cy += 1; // Cursor down
                else if (E.rowOff < E.numRows) E.rowOff += 1; // Scroll down

                int nextLineLen = E.rows[E.rowOff + E.cy]->len;
                if (nextLineLen < E.cx) E.cx = nextLineLen;
                else E.cx = nextLineLen < E.colOff ? nextLineLen : E.colOff;
            }
            break;
        case ARROW_RIGHT:
            if (E.cx < E.screenCols - 1 &&
                    E.cx < E.rows[E.rowOff + E.cy]->len) {
                E.cx += 1;
                E.colOff = E.cx;
            }
            break;
        case HOME_KEY:
            E.cx = 0;
            break;
        case END_KEY:
            E.cx = E.screenCols - 1;
            break;
        default:
            die("editMoveCursor");
    }
}

void editorProcessKeyPress() {
    int c = editorReadKey();

    switch (c) {
        case CTRL_KEY('q'):
            write(STDOUT_FILENO, "\x1b[2J", 4);
            write(STDOUT_FILENO, "\x1b[H", 3);
            exit(1);
            break;

        case ARROW_UP:
        case ARROW_DOWN:
        case ARROW_LEFT:
        case ARROW_RIGHT:
        case HOME_KEY:
        case END_KEY:
            editMoveCursor(c);
            break;

        case PAGE_UP:
        case PAGE_DOWN:
            {
                int times = E.screenRows;
                while (times--)
                    editMoveCursor(c == PAGE_UP ? ARROW_UP : ARROW_DOWN);
            }
            break;
    }
}

/** Output **/

void printWelcome(struct abuf* ab) {
    char welcome[80];
    int welcomeLen = snprintf(welcome, sizeof(welcome), " Kilo editor -- version %s", KILO_VERSION);
    if (welcomeLen > E.screenCols) {
        welcomeLen = E.screenCols;
    } else {
        int padding = (E.screenCols - welcomeLen) / 2;
        while (padding--)
            abAppend(ab, " ", 1);
    }
    abAppend(ab, welcome, welcomeLen);
}

void editorDrawRows(struct abuf* ab) {
    for (int y = 0; y < E.screenRows; y++) {
        int fileRow = y + E.rowOff;
        if (fileRow < E.numRows) {
            int len = E.rows[fileRow]->len;
            if (len > E.screenCols)
                len = E.screenCols;
            abAppend(ab, E.rows[fileRow]->line, len);
        } else {
            abAppend(ab, "~", 1);
            if (E.numRows == 0 && y == E.screenRows / 3) {
                printWelcome(ab);
            }
        }
        abAppend(ab, "\x1b[K", 3);
        if (y < E.screenRows - 1)
            abAppend(ab, "\r\n", 2);
    }
}

void editorRefreshScreen() {
    struct abuf ab = ABUF_INIT;

    // Reposition the cursor
    // H accepts row;col args
    abAppend(&ab, "\x1b[H", 3);
    editorDrawRows(&ab);

    char nbuf[32];
    snprintf(nbuf, sizeof(nbuf), "\x1b[%d;%dH", E.cy + 1, E.cx + 1);
    abAppend(&ab, nbuf, strlen(nbuf));

    write(STDOUT_FILENO, ab.b, ab.len);
    abFree(&ab);
}

/** File **/

void editorAppendRow(char* line, int len) {
    E.rows = realloc(E.rows, sizeof(erow*) * (E.numRows + 1));

    erow *row = malloc(len + 1);
    row->len = len;
    row->line = malloc(len + 1);
    memcpy(row->line, line, len);
    row->line[len] = '\0';

    E.rows[E.numRows] = row;
    E.numRows += 1;
}

void editorOpen(char* filename) {
    FILE* fp = fopen(filename, "r");
    if (!fp) die("fopen");

    char* line = NULL;
    size_t lineCap = 0;
    ssize_t lineLen;

    while((lineLen = getline(&line, &lineCap, fp)) != -1) {
        if (lineLen != -1)
            while (lineLen > 0 && (
                        line[lineLen - 1] == '\n' ||
                        line[lineLen - 1] == '\r'))
                lineLen--;
        editorAppendRow(line, lineLen);
    }
    free(line);
    fclose(fp);
}

/** Init **/

void initEditor() {
    E.cx = 0;
    E.cy = 0;
    E.numRows = 0;
    E.rows = NULL;
    E.rowOff = 0;
    E.colOff = 0;

    if (getWindowSize(&E.screenRows, &E.screenCols) == -1)
        die("getWindowSize");
}

int main(int argc, char* argv[]) {
    enableRawMode();
    // TODO: What if window is resized?
    initEditor();

    if (argc >= 2)
        editorOpen(argv[1]);

    while (1) {
        editorRefreshScreen();
        editorProcessKeyPress();
    }
    return 0;
}
