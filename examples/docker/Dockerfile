FROM ubuntu:24.04

# Install the cartage binary and create compat symlinks
COPY cartage /usr/local/bin/cartage
RUN ln -s cartage /usr/local/bin/notify-send && \
    ln -s cartage /usr/local/bin/xdg-open && \
    ln -s cartage /usr/local/bin/pbcopy && \
    ln -s cartage /usr/local/bin/pbpaste && \
    # Add a dialog tool if you need blocking alerts/confirmations (pick one)
    ln -s cartage /usr/local/bin/yad

CMD ["bash"]
