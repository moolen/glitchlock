#include <security/pam_appl.h>
#include <security/pam_misc.h>
#include <stdio.h>
#include <pwd.h>

const char* test_password = NULL;

const char* get_current_user() {
    struct passwd *pw;
    const char *user;
    if ((pw = getpwuid(getuid())) == NULL)
        return NULL;
    if ((user = pw->pw_name) == NULL)
        return NULL;
    return user;
}

static int conv_callback(int num_msg, const struct pam_message **msg, struct pam_response **resp, void *appdata_ptr) {
    if (num_msg == 0)
        return 1;

    if ((*resp = calloc(num_msg, sizeof(struct pam_response))) == NULL) {
        return 1;
    }

    for (int c = 0; c < num_msg; c++) {
        if (msg[c]->msg_style != PAM_PROMPT_ECHO_OFF &&
            msg[c]->msg_style != PAM_PROMPT_ECHO_ON)
            continue;

        resp[c]->resp_retcode = 0;
        if ((resp[c]->resp = strdup(test_password)) == NULL) {
            return 1;
        }
    }

    return 0;
}

int check_password(const char* user, const char* password) {
    pam_handle_t *pam_handle=NULL;
    int retval;
    struct passwd *pw;
    struct pam_conv conv = {
        conv_callback,
        NULL
    };
    test_password = password;

    retval = pam_start("login", user, &conv, &pam_handle);

    if (retval == PAM_SUCCESS)
        retval = pam_authenticate(pam_handle, 0);

    if (retval == PAM_SUCCESS)
        retval = pam_acct_mgmt(pam_handle, 0);

    if (pam_end(pam_handle,retval) != PAM_SUCCESS) {
        pam_handle = NULL;
    }
    test_password = NULL;
    return ( retval == PAM_SUCCESS ? 1:0 );
}

int check_current_user(const char* password) {
    const char* user;
    if ((user = get_current_user()) == NULL)
        return 0;
    return check_password(user, password);
}
